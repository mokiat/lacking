package game

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/async"
)

type ModelDefinition struct {
	nodes                     []nodeDefinition
	armatures                 []armatureDefinition
	animations                []*AnimationDefinition
	textures                  []render.Texture
	materialDefinitions       []*graphics.Material
	meshDefinitions           []*graphics.MeshDefinition
	meshInstances             []meshInstance
	bodyDefinitions           []*physics.BodyDefinition
	bodyInstances             []bodyInstance
	pointLightInstances       []pointLightInstance
	spotLightInstances        []spotLightInstance
	directionalLightInstances []directionalLightInstance
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

type pointLightInstance struct {
	Name      string
	NodeIndex int
	EmitRange float64
	EmitColor dprec.Vec3
}

type spotLightInstance struct {
	Name               string
	NodeIndex          int
	EmitRange          float64
	EmitOuterConeAngle dprec.Angle
	EmitInnerConeAngle dprec.Angle
	EmitColor          dprec.Vec3
}

type directionalLightInstance struct {
	Name      string
	NodeIndex int
	EmitRange float64
	EmitColor dprec.Vec3
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
				AerodynamicShapes:      nil, // TODO
				CollisionSpheres:       r.constructCollisionSpheres(definitionAsset),
				CollisionBoxes:         r.constructCollisionBoxes(definitionAsset),
				CollisionMeshes:        r.constructCollisionMeshes(definitionAsset),
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

	textures := make([]render.Texture, len(modelAsset.Textures))
	for i, textureAsset := range modelAsset.Textures {
		textures[i] = r.allocateTwoDTexture(&textureAsset)
	}

	materialDefinitions := make([]*graphics.Material, len(modelAsset.Materials))
	for i, materialAsset := range modelAsset.Materials {
		pbrAsset := asset.NewPBRMaterialView(&materialAsset)

		var albedoTexture render.Texture
		if ref := pbrAsset.BaseColorTexture(); ref.Valid() {
			if ref.TextureIndex >= 0 {
				albedoTexture = textures[ref.TextureIndex]
			} else {
				var promise async.Promise[render.Texture]
				r.gfxWorker.ScheduleVoid(func() {
					promise = r.OpenTwoDTexture(ref.TextureID)
				}).Wait()
				texture, err := promise.Wait()
				if err != nil {
					return nil, fmt.Errorf("error loading albedo texture: %w", err)
				}
				albedoTexture = texture
			}
		}

		var metallicRoughnessTexture render.Texture
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

		var normalTexture render.Texture
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
			materialDefinitions[i] = gfxEngine.CreatePBRMaterial(graphics.PBRMaterialInfo{
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

	meshGeometries := make([]*graphics.MeshGeometry, len(modelAsset.MeshDefinitions))
	for i, definitionAsset := range modelAsset.MeshDefinitions {
		meshFragments := make([]graphics.MeshGeometryFragmentInfo, len(definitionAsset.Fragments))
		for j, fragmentAsset := range definitionAsset.Fragments {
			material := materialDefinitions[fragmentAsset.MaterialIndex]
			meshFragments[j] = graphics.MeshGeometryFragmentInfo{
				Name:            material.Name(),
				Topology:        resolveTopology(fragmentAsset.Topology),
				IndexByteOffset: fragmentAsset.IndexOffset,
				IndexCount:      fragmentAsset.IndexCount,
			}
		}

		gfxEngine := r.engine.Graphics()
		r.gfxWorker.ScheduleVoid(func() {
			meshGeometries[i] = gfxEngine.CreateMeshGeometry(graphics.MeshGeometryInfo{
				VertexData:           definitionAsset.VertexData,
				VertexFormat:         resolveVertexFormat(definitionAsset.VertexLayout),
				IndexData:            definitionAsset.IndexData,
				IndexFormat:          resolveIndexFormat(definitionAsset.IndexLayout),
				Fragments:            meshFragments,
				BoundingSphereRadius: definitionAsset.BoundingSphereRadius,
			})
		}).Wait()
	}

	meshDefinitions := make([]*graphics.MeshDefinition, len(modelAsset.MeshDefinitions))
	for i, definitionAsset := range modelAsset.MeshDefinitions {
		gfxEngine := r.engine.Graphics()
		r.gfxWorker.ScheduleVoid(func() {
			var materials []*graphics.Material
			for _, fragmentAsset := range definitionAsset.Fragments {
				materials = append(materials, materialDefinitions[fragmentAsset.MaterialIndex])
			}
			meshDefinitions[i] = gfxEngine.CreateMeshDefinition(graphics.MeshDefinitionInfo{
				Geometry:  meshGeometries[i],
				Materials: materials,
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

	pointLightInstances := make([]pointLightInstance, 0)
	spotLightInstances := make([]spotLightInstance, 0)
	directionalLightInstances := make([]directionalLightInstance, 0)
	for _, instanceAsset := range modelAsset.LightInstances {
		switch instanceAsset.Type {
		case asset.LightTypePoint:
			pointLightInstances = append(pointLightInstances, pointLightInstance{
				Name:      instanceAsset.Name,
				NodeIndex: int(instanceAsset.NodeIndex),
				EmitRange: instanceAsset.EmitRange,
				EmitColor: instanceAsset.EmitColor,
			})
		case asset.LightTypeSpot:
			spotLightInstances = append(spotLightInstances, spotLightInstance{
				Name:               instanceAsset.Name,
				NodeIndex:          int(instanceAsset.NodeIndex),
				EmitRange:          instanceAsset.EmitRange,
				EmitOuterConeAngle: instanceAsset.EmitOuterConeAngle,
				EmitInnerConeAngle: instanceAsset.EmitInnerConeAngle,
				EmitColor:          instanceAsset.EmitColor,
			})
		case asset.LightTypeDirectional:
			directionalLightInstances = append(directionalLightInstances, directionalLightInstance{
				Name:      instanceAsset.Name,
				NodeIndex: int(instanceAsset.NodeIndex),
				EmitRange: instanceAsset.EmitRange,
				EmitColor: instanceAsset.EmitColor,
			})
		default:
			return nil, fmt.Errorf("unknown light type: %d", instanceAsset.Type)
		}
	}

	return &ModelDefinition{
		nodes:                     nodes,
		armatures:                 armatures,
		animations:                animations,
		textures:                  textures,
		materialDefinitions:       materialDefinitions,
		meshDefinitions:           meshDefinitions,
		meshInstances:             meshInstances,
		bodyDefinitions:           bodyDefinitions,
		bodyInstances:             bodyInstances,
		pointLightInstances:       pointLightInstances,
		spotLightInstances:        spotLightInstances,
		directionalLightInstances: directionalLightInstances,
	}, nil
}

func (r *ResourceSet) releaseModel(model *ModelDefinition) {
	for _, texture := range model.textures {
		texture.Release()
	}
	model.textures = nil
}

func (r *ResourceSet) constructCollisionSpheres(bodyDef asset.BodyDefinition) []collision.Sphere {
	result := make([]collision.Sphere, len(bodyDef.CollisionSpheres))
	for i, collisionSphereAsset := range bodyDef.CollisionSpheres {
		result[i] = collision.NewSphere(
			collisionSphereAsset.Translation,
			collisionSphereAsset.Radius,
		)
	}
	return result
}

func (r *ResourceSet) constructCollisionBoxes(bodyDef asset.BodyDefinition) []collision.Box {
	result := make([]collision.Box, len(bodyDef.CollisionBoxes))
	for i, collisionBoxAsset := range bodyDef.CollisionBoxes {
		result[i] = collision.NewBox(
			collisionBoxAsset.Translation,
			collisionBoxAsset.Rotation,
			dprec.NewVec3(collisionBoxAsset.Width, collisionBoxAsset.Height, collisionBoxAsset.Lenght),
		)
	}
	return result
}

func (r *ResourceSet) constructCollisionMeshes(bodyDef asset.BodyDefinition) []collision.Mesh {
	result := make([]collision.Mesh, len(bodyDef.CollisionMeshes))
	for i, collisionMeshAsset := range bodyDef.CollisionMeshes {
		transform := collision.TRTransform(collisionMeshAsset.Translation, collisionMeshAsset.Rotation)
		triangles := make([]collision.Triangle, len(collisionMeshAsset.Triangles))
		for j, triangleAsset := range collisionMeshAsset.Triangles {
			template := collision.NewTriangle(
				triangleAsset.A,
				triangleAsset.B,
				triangleAsset.C,
			)
			triangles[j].Replace(template, transform)
		}
		result[i] = collision.NewMesh(triangles)
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

func resolveIndexFormat(layout asset.IndexLayout) render.IndexFormat {
	switch layout {
	case asset.IndexLayoutUint16:
		return render.IndexFormatUnsignedShort
	case asset.IndexLayoutUint32:
		return render.IndexFormatUnsignedInt
	default:
		panic(fmt.Errorf("unsupported index layout: %d", layout))
	}
}

func resolveTopology(primitive asset.MeshTopology) render.Topology {
	switch primitive {
	case asset.MeshTopologyPoints:
		return render.TopologyPoints
	case asset.MeshTopologyLines:
		return render.TopologyLineList
	case asset.MeshTopologyLineStrip:
		return render.TopologyLineStrip
	case asset.MeshTopologyTriangles:
		return render.TopologyTriangleList
	case asset.MeshTopologyTriangleStrip:
		return render.TopologyTriangleStrip
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
	root       *hierarchy.Node

	nodes                     []*hierarchy.Node
	armatures                 []*graphics.Armature
	animations                []*Animation
	bodyInstances             []physics.Body
	pointLightInstances       []*graphics.PointLight
	spotLightInstances        []*graphics.SpotLight
	directionalLightInstances []*graphics.DirectionalLight
}

func (m *Model) Root() *hierarchy.Node {
	return m.root
}

func (m *Model) FindNode(name string) *hierarchy.Node {
	return m.root.FindNode(name)
}

func (m *Model) BodyInstances() []physics.Body {
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
