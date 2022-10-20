package game

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/util/async"
	"github.com/mokiat/lacking/util/shape"
)

type ModelDefinition struct {
	nodes               []nodeDefinition
	armatures           []armatureDefinition
	animations          []*AnimationDefinition
	textures            []*TwoDTexture
	materialDefinitions []*graphics.MaterialDefinition
	meshDefinitions     []*graphics.MeshDefinition
	meshInstances       []meshInstance
	bodyDefinitions     []*physics.BodyDefinition
	bodyInstances       []bodyInstance
}

func (d *ModelDefinition) FindAnimation(name string) *AnimationDefinition {
	for _, def := range d.animations {
		if def.name == name {
			return def
		}
	}
	return nil
}

type nodeDefinition struct {
	ParentIndex int
	Name        string
	Position    dprec.Vec3
	Rotation    dprec.Quat
	Scale       dprec.Vec3
}

type meshInstance struct {
	Name            string
	NodeIndex       int
	ArmatureIndex   int
	DefinitionIndex int
}

type bodyInstance struct {
	Name            string
	NodeIndex       int
	DefinitionIndex int
}

type armatureDefinition struct {
	Joints []armatureJoint
}

func (d armatureDefinition) InverseBindMatrices() []sprec.Mat4 {
	result := make([]sprec.Mat4, len(d.Joints))
	for i, joint := range d.Joints {
		result[i] = joint.InverseBindMatrix
	}
	return result
}

type armatureJoint struct {
	NodeIndex         int
	InverseBindMatrix sprec.Mat4
}

func (r *ResourceSet) loadModel(resource asset.Resource) (*ModelDefinition, error) {
	modelAsset := new(asset.Model)
	ioTask := func() error {
		return resource.ReadContent(modelAsset)
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	return r.allocateModel(modelAsset)
}

func (r *ResourceSet) allocateModel(modelAsset *asset.Model) (*ModelDefinition, error) {
	nodes := make([]nodeDefinition, len(modelAsset.Nodes))
	for i, nodeAsset := range modelAsset.Nodes {
		nodes[i] = nodeDefinition{
			ParentIndex: int(nodeAsset.ParentIndex),
			Name:        nodeAsset.Name,
			Position:    nodeAsset.Translation,
			Rotation:    nodeAsset.Rotation,
			Scale:       nodeAsset.Scale,
		}
	}

	armatures := make([]armatureDefinition, len(modelAsset.Armatures))
	for i, armatureAsset := range modelAsset.Armatures {
		joints := make([]armatureJoint, len(armatureAsset.Joints))
		for j, jointAsset := range armatureAsset.Joints {
			joints[j] = armatureJoint{
				NodeIndex:         int(jointAsset.NodeIndex),
				InverseBindMatrix: jointAsset.InverseBindMatrix,
			}
		}
		armatures[i] = armatureDefinition{
			Joints: joints,
		}
	}

	animations := make([]*AnimationDefinition, len(modelAsset.Animations))
	for i, animationAsset := range modelAsset.Animations {
		bindings := make([]AnimationBindingDefinitionInfo, len(animationAsset.Bindings))
		for j, assetBinding := range animationAsset.Bindings {
			translationKeyframes := make([]Keyframe[dprec.Vec3], len(assetBinding.TranslationKeyframes))
			for k, keyframe := range assetBinding.TranslationKeyframes {
				translationKeyframes[k] = Keyframe[dprec.Vec3]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Translation,
				}
			}
			rotationKeyframes := make([]Keyframe[dprec.Quat], len(assetBinding.RotationKeyframes))
			for k, keyframe := range assetBinding.RotationKeyframes {
				rotationKeyframes[k] = Keyframe[dprec.Quat]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Rotation,
				}
			}
			scaleKeyframes := make([]Keyframe[dprec.Vec3], len(assetBinding.ScaleKeyframes))
			for k, keyframe := range assetBinding.ScaleKeyframes {
				scaleKeyframes[k] = Keyframe[dprec.Vec3]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Scale,
				}
			}
			bindings[j] = AnimationBindingDefinitionInfo{
				NodeIndex:            int(assetBinding.NodeIndex),
				NodeName:             assetBinding.NodeName,
				TranslationKeyframes: translationKeyframes,
				RotationKeyframes:    rotationKeyframes,
				ScaleKeyframes:       scaleKeyframes,
			}
		}
		r.gfxWorker.ScheduleVoid(func() {
			animations[i] = r.engine.CreateAnimationDefinition(AnimationDefinitionInfo{
				Name:      animationAsset.Name,
				StartTime: animationAsset.StartTime,
				EndTime:   animationAsset.EndTime,
				Bindings:  bindings,
			})
		}).Wait()
	}

	bodyDefinitions := make([]*physics.BodyDefinition, len(modelAsset.BodyDefinitions))
	for i, definitionAsset := range modelAsset.BodyDefinitions {
		physicsEngine := r.engine.Physics()
		r.gfxWorker.ScheduleVoid(func() {
			bodyDefinitions[i] = physicsEngine.CreateBodyDefinition(physics.BodyDefinitionInfo{
				Mass:                   definitionAsset.Mass,
				MomentOfInertia:        definitionAsset.MomentOfInertia,
				RestitutionCoefficient: definitionAsset.RestitutionCoefficient,
				DragFactor:             definitionAsset.DragFactor,
				AngularDragFactor:      definitionAsset.AngularDragFactor,
				CollisionShapes:        r.constructCollisionShapes(definitionAsset),
				AerodynamicShapes:      nil, // TODO
			})
		}).Wait()
	}

	bodyInstances := make([]bodyInstance, len(modelAsset.BodyInstances))
	for i, instanceAsset := range modelAsset.BodyInstances {
		bodyInstances[i] = bodyInstance{
			Name:            instanceAsset.Name,
			NodeIndex:       int(instanceAsset.NodeIndex),
			DefinitionIndex: int(instanceAsset.BodyIndex),
		}
	}

	textures := make([]*TwoDTexture, len(modelAsset.Textures))
	for i, textureAsset := range modelAsset.Textures {
		textures[i] = r.allocateTwoDTexture(&textureAsset)
	}

	materialDefinitions := make([]*graphics.MaterialDefinition, len(modelAsset.Materials))
	for i, materialAsset := range modelAsset.Materials {
		pbrAsset := asset.NewPBRMaterialView(&materialAsset)

		var albedoTexture *graphics.TwoDTexture
		if ref := pbrAsset.BaseColorTexture(); ref.Valid() {
			if ref.TextureIndex >= 0 {
				albedoTexture = textures[ref.TextureIndex].gfxTexture
			} else {
				var promise async.Promise[*TwoDTexture]
				r.gfxWorker.ScheduleVoid(func() {
					promise = r.OpenTwoDTexture(ref.TextureID)
				}).Wait()
				texture, err := promise.Wait()
				if err != nil {
					return nil, fmt.Errorf("error loading albedo texture: %w", err)
				}
				albedoTexture = texture.gfxTexture
			}
		}

		var metallicRoughnessTexture *graphics.TwoDTexture
		// if texID := pbrAsset.MetallicRoughnessTexture(); texID != "" {
		// 	var promise async.Promise[*TwoDTexture]
		// 	r.gfxWorker.ScheduleVoid(func() {
		// 		promise = r.OpenTwoDTexture(texID)
		// 	}).Wait()
		// 	texture, err := promise.Wait()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("error loading albedo texture: %w", err)
		// 	}
		// 	metallicRoughnessTexture = texture.gfxTexture
		// }

		var normalTexture *graphics.TwoDTexture
		// if texID := pbrAsset.NormalTexture(); texID != "" {
		// 	var promise async.Promise[*TwoDTexture]
		// 	r.gfxWorker.ScheduleVoid(func() {
		// 		promise = r.OpenTwoDTexture(texID)
		// 	}).Wait()
		// 	texture, err := promise.Wait()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("error loading albedo texture: %w", err)
		// 	}
		// 	normalTexture = texture.gfxTexture
		// }

		gfxEngine := r.engine.Graphics()
		r.gfxWorker.ScheduleVoid(func() {
			materialDefinitions[i] = gfxEngine.CreatePBRMaterialDefinition(graphics.PBRMaterialInfo{
				BackfaceCulling:          materialAsset.BackfaceCulling,
				AlphaBlending:            materialAsset.Blending,
				AlphaTesting:             materialAsset.AlphaTesting,
				AlphaThreshold:           materialAsset.AlphaThreshold,
				Metallic:                 pbrAsset.Metallic(),
				Roughness:                pbrAsset.Roughness(),
				MetallicRoughnessTexture: metallicRoughnessTexture,
				AlbedoColor:              pbrAsset.BaseColor(),
				AlbedoTexture:            albedoTexture,
				NormalScale:              pbrAsset.NormalScale(),
				NormalTexture:            normalTexture,
			})
		}).Wait()
	}

	meshDefinitions := make([]*graphics.MeshDefinition, len(modelAsset.MeshDefinitions))
	for i, definitionAsset := range modelAsset.MeshDefinitions {
		meshFragments := make([]graphics.MeshFragmentDefinitionInfo, len(definitionAsset.Fragments))
		for j, fragmentAsset := range definitionAsset.Fragments {
			material := materialDefinitions[fragmentAsset.MaterialIndex]
			meshFragments[j] = graphics.MeshFragmentDefinitionInfo{
				Primitive:   resolvePrimitive(fragmentAsset.Topology),
				IndexOffset: int(fragmentAsset.IndexOffset),
				IndexCount:  int(fragmentAsset.IndexCount),
				Material:    material,
			}
		}

		gfxEngine := r.engine.Graphics()
		r.gfxWorker.ScheduleVoid(func() {
			meshDefinitions[i] = gfxEngine.CreateMeshDefinition(graphics.MeshDefinitionInfo{
				VertexData:           definitionAsset.VertexData,
				VertexFormat:         resolveVertexFormat(definitionAsset.VertexLayout),
				IndexData:            definitionAsset.IndexData,
				IndexFormat:          resolveIndexFormat(definitionAsset.IndexLayout),
				Fragments:            meshFragments,
				BoundingSphereRadius: definitionAsset.BoundingSphereRadius,
			})
		}).Wait()
	}

	meshInstances := make([]meshInstance, len(modelAsset.MeshInstances))
	for i, instanceAsset := range modelAsset.MeshInstances {
		meshInstances[i] = meshInstance{
			Name:            instanceAsset.Name,
			NodeIndex:       int(instanceAsset.NodeIndex),
			ArmatureIndex:   int(instanceAsset.ArmatureIndex),
			DefinitionIndex: int(instanceAsset.DefinitionIndex),
		}
	}

	return &ModelDefinition{
		nodes:               nodes,
		armatures:           armatures,
		animations:          animations,
		textures:            textures,
		materialDefinitions: materialDefinitions,
		meshDefinitions:     meshDefinitions,
		meshInstances:       meshInstances,
		bodyDefinitions:     bodyDefinitions,
		bodyInstances:       bodyInstances,
	}, nil
}

func (r *ResourceSet) releaseModel(model *ModelDefinition) {
	for _, texture := range model.textures {
		r.releaseTwoDTexture(texture)
	}
	model.textures = nil
}

func (r *ResourceSet) constructCollisionShapes(bodyDef asset.BodyDefinition) []physics.CollisionShape {
	var result []physics.CollisionShape
	for _, collisionBoxAsset := range bodyDef.CollisionBoxes {
		result = append(result,
			shape.NewStaticBox(
				collisionBoxAsset.Width,
				collisionBoxAsset.Height,
				collisionBoxAsset.Lenght,
			).WithTransform(shape.NewTransform(
				collisionBoxAsset.Translation,
				collisionBoxAsset.Rotation,
			)),
		)
	}
	for _, collisionSphereAsset := range bodyDef.CollisionSpheres {
		result = append(result,
			shape.NewStaticSphere(
				collisionSphereAsset.Radius,
			).WithTransform(shape.NewTransform(
				collisionSphereAsset.Translation,
				collisionSphereAsset.Rotation,
			)),
		)
	}
	for _, collisionMeshAsset := range bodyDef.CollisionMeshes {
		triangles := make([]shape.StaticTriangle, len(collisionMeshAsset.Triangles))
		for j, triangleAsset := range collisionMeshAsset.Triangles {
			triangles[j] = shape.NewStaticTriangle(
				shape.Point(triangleAsset.A),
				shape.Point(triangleAsset.B),
				shape.Point(triangleAsset.C),
			)
		}
		result = append(result,
			shape.NewStaticMesh(
				triangles,
			).WithTransform(shape.NewTransform(
				collisionMeshAsset.Translation,
				collisionMeshAsset.Rotation,
			)),
		)
	}
	return result
}

func resolveVertexFormat(layout asset.VertexLayout) graphics.VertexFormat {
	return graphics.VertexFormat{
		HasCoord:            layout.CoordOffset != asset.UnspecifiedOffset,
		CoordOffsetBytes:    int(layout.CoordOffset),
		CoordStrideBytes:    int(layout.CoordStride),
		HasNormal:           layout.NormalOffset != asset.UnspecifiedOffset,
		NormalOffsetBytes:   int(layout.NormalOffset),
		NormalStrideBytes:   int(layout.NormalStride),
		HasTangent:          layout.TangentOffset != asset.UnspecifiedOffset,
		TangentOffsetBytes:  int(layout.TangentOffset),
		TangentStrideBytes:  int(layout.TangentStride),
		HasTexCoord:         layout.TexCoordOffset != asset.UnspecifiedOffset,
		TexCoordOffsetBytes: int(layout.TexCoordOffset),
		TexCoordStrideBytes: int(layout.TexCoordStride),
		HasColor:            layout.ColorOffset != asset.UnspecifiedOffset,
		ColorOffsetBytes:    int(layout.ColorOffset),
		ColorStrideBytes:    int(layout.ColorStride),
		HasWeights:          layout.WeightsOffset != asset.UnspecifiedOffset,
		WeightsOffsetBytes:  int(layout.WeightsOffset),
		WeightsStrideBytes:  int(layout.WeightsStride),
		HasJoints:           layout.JointsOffset != asset.UnspecifiedOffset,
		JointsOffsetBytes:   int(layout.JointsOffset),
		JointsStrideBytes:   int(layout.JointsStride),
	}
}

func resolveIndexFormat(layout asset.IndexLayout) graphics.IndexFormat {
	switch layout {
	case asset.IndexLayoutUint16:
		return graphics.IndexFormatU16
	case asset.IndexLayoutUint32:
		return graphics.IndexFormatU32
	default:
		panic(fmt.Errorf("unsupported index layout: %d", layout))
	}
}

func resolvePrimitive(primitive asset.MeshTopology) graphics.Primitive {
	switch primitive {
	case asset.MeshTopologyPoints:
		return graphics.PrimitivePoints
	case asset.MeshTopologyLines:
		return graphics.PrimitiveLines
	case asset.MeshTopologyLineStrip:
		return graphics.PrimitiveLineStrip
	case asset.MeshTopologyLineLoop:
		return graphics.PrimitiveLineLoop
	case asset.MeshTopologyTriangles:
		return graphics.PrimitiveTriangles
	case asset.MeshTopologyTriangleStrip:
		return graphics.PrimitiveTriangleStrip
	case asset.MeshTopologyTriangleFan:
		return graphics.PrimitiveTriangleFan
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}

// ModelInfo contains the information necessary to place a Model
// instance into a Scene.
type ModelInfo struct {
	// Name specifies the name of this instance. This should not be
	// confused with the name of the definition.
	Name string

	// Definition specifies the template from which this instance will
	// be created.
	Definition *ModelDefinition

	// Position is used to specify a location for the model instance.
	Position dprec.Vec3

	// Rotation is used to specify a rotation for the model instance.
	Rotation dprec.Quat

	// Scale is used to specify a scale for the model instance.
	Scale dprec.Vec3

	// IsDynamic determines whether the model can be repositioned once
	// placed in the Scene.
	// (i.e. whether it should be added to the scene hierarchy)
	IsDynamic bool

	// PrepareAnimations indicates whether animation definitions should be
	// instantiated for this model.
	PrepareAnimations bool
}

type Model struct {
	definition *ModelDefinition
	root       *Node

	nodes         []*Node
	armatures     []*graphics.Armature
	animations    []*Animation
	bodyInstances []*physics.Body
}

func (m *Model) Root() *Node {
	return m.root
}

func (m *Model) BodyInstances() []*physics.Body {
	return m.bodyInstances
}

func (m *Model) Animations() []*Animation {
	return m.animations
}

func (m *Model) FindAnimation(name string) *Animation {
	for _, animation := range m.animations {
		if animation.name == name {
			return animation
		}
	}
	return nil
}
