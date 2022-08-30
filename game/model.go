package game

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/util/shape"
)

type NodeDefinition struct {
	ParentIndex int
	Name        string
	Position    dprec.Vec3
	Rotation    dprec.Quat
	Scale       dprec.Vec3
}

type ArmatureDefinition struct {
	GraphicsTemplate *graphics.ArmatureTemplate
}

type MaterialDefinition struct {
}

type ModelDefinition struct {
	nodes           []NodeDefinition
	meshDefinitions []*graphics.MeshDefinition
	meshInstances   []meshInstance
	bodyDefinitions []*physics.BodyDefinition
	bodyInstances   []bodyInstance

	// TODO: Fix these as well
	Animations []*AnimationDefinition
	Armatures  []*ArmatureDefinition
	Materials  []*MaterialDefinition
}

type meshInstance struct {
	Name            string
	NodeIndex       int
	DefinitionIndex int
	// Armature        *ArmatureDefinition
}

type bodyInstance struct {
	Name            string
	NodeIndex       int
	DefinitionIndex int
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
}

type Model struct {
	definition *ModelDefinition
	root       *Node

	nodes     []*Node
	armatures []*graphics.Armature
	materials []*graphics.Material
}

func (r *ResourceSet) allocateModel(resource asset.Resource) (*ModelDefinition, error) {
	modelAsset := new(asset.Model)
	if err := resource.ReadContent(modelAsset); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	result := &ModelDefinition{}

	nodes := make([]NodeDefinition, len(modelAsset.Nodes))
	for i, nodeAsset := range modelAsset.Nodes {
		nodes[i] = NodeDefinition{
			ParentIndex: int(nodeAsset.ParentIndex),
			Name:        nodeAsset.Name,
			Position:    nodeAsset.Translation,
			Rotation:    nodeAsset.Rotation,
			Scale:       nodeAsset.Scale,
		}
	}
	result.nodes = nodes

	bodyDefinitions := make([]*physics.BodyDefinition, len(modelAsset.BodyDefinitions))
	for i, definitionAsset := range modelAsset.BodyDefinitions {
		physicsEngine := r.engine.Physics()
		r.gfxWorker.Schedule(func() {
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
	result.bodyDefinitions = bodyDefinitions

	bodyInstances := make([]bodyInstance, len(modelAsset.BodyInstances))
	for i, instanceAsset := range modelAsset.BodyInstances {
		bodyInstances[i] = bodyInstance{
			Name:            instanceAsset.Name,
			NodeIndex:       int(instanceAsset.NodeIndex),
			DefinitionIndex: int(instanceAsset.BodyIndex),
		}
	}
	result.bodyInstances = bodyInstances

	materialDefinitions := make([]*graphics.MaterialDefinition, len(modelAsset.Materials))
	for i, materialAsset := range modelAsset.Materials {
		pbrAsset := asset.NewPBRMaterialView(&materialAsset)

		// TODO
		// var albedoTexture *graphics.TwoDTexture
		// if texID := pbrAsset.BaseColorTexture(); texID != "" {
		// 	var placeholder Placeholder[*TwoDTexture]
		// 	r.gfxWorker.Schedule(func() {
		// 		placeholder = r.OpenTwoDTexture(texID)
		// 	})
		// 	texture, err := placeholder.Wait()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("error loading albedo texture: %w", err)
		// 	}
		// 	albedoTexture = texture.gfxTexture
		// }

		// var metallicRoughnessTexture *graphics.TwoDTexture
		// if texID := assetPBR.MetallicRoughnessTexture(); texID != "" {
		// 	var placeholder Placeholder[*TwoDTexture]

		// 	texture, err := placeholder.Wait()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("error loading albedo texture: %w", err)
		// 	}

		// 	metallicRoughnessTexture = texture
		// }

		gfxEngine := r.engine.Graphics()
		r.gfxWorker.Schedule(func() {
			materialDefinitions[i] = gfxEngine.CreatePBRMaterialDefinition(graphics.PBRMaterialInfo{
				BackfaceCulling:          materialAsset.BackfaceCulling,
				AlphaBlending:            materialAsset.Blending,
				AlphaTesting:             materialAsset.AlphaTesting,
				AlphaThreshold:           materialAsset.AlphaThreshold,
				Metallic:                 pbrAsset.Metallic(),
				Roughness:                pbrAsset.Roughness(),
				MetallicRoughnessTexture: nil, // FIXME
				AlbedoColor:              pbrAsset.BaseColor(),
				AlbedoTexture:            nil, // FIXME
				NormalScale:              pbrAsset.NormalScale(),
				NormalTexture:            nil, // FIXME
			})
		}).Wait()
	}

	meshDefinitions := make([]*graphics.MeshDefinition, len(modelAsset.BodyDefinitions))
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
		r.gfxWorker.Schedule(func() {
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

	result.meshDefinitions = meshDefinitions

	meshInstances := make([]meshInstance, len(modelAsset.MeshInstances))
	for i, instanceAsset := range modelAsset.MeshInstances {
		meshInstances[i] = meshInstance{
			Name:            instanceAsset.Name,
			NodeIndex:       int(instanceAsset.NodeIndex),
			DefinitionIndex: int(instanceAsset.DefinitionIndex),
			// TODO: Armature
		}
	}
	result.meshInstances = meshInstances

	return result, nil
}

func (r *ResourceSet) releaseModel(model *ModelDefinition) {
	// TODO
}

func (r *ResourceSet) constructCollisionShapes(bodyDef asset.BodyDefinition) []physics.CollisionShape {
	var result []physics.CollisionShape
	for _, collisionBoxAsset := range bodyDef.CollisionBoxes {
		result = append(result, shape.NewPlacement(
			shape.NewStaticBox(
				collisionBoxAsset.Width,
				collisionBoxAsset.Height,
				collisionBoxAsset.Lenght,
			),
			collisionBoxAsset.Translation,
			collisionBoxAsset.Rotation,
		))
	}
	for _, collisionSphereAsset := range bodyDef.CollisionSpheres {
		result = append(result, shape.NewPlacement(
			shape.NewStaticSphere(
				collisionSphereAsset.Radius,
			),
			collisionSphereAsset.Translation,
			collisionSphereAsset.Rotation,
		))
	}
	for _, collisionMeshAsset := range bodyDef.CollisionMeshes {
		triangles := make([]shape.StaticTriangle, len(collisionMeshAsset.Triangles))
		for j, triangleAsset := range collisionMeshAsset.Triangles {
			triangles[j] = shape.NewStaticTriangle(
				triangleAsset.A,
				triangleAsset.B,
				triangleAsset.C,
			)
		}
		result = append(result, shape.NewPlacement(
			shape.NewStaticMesh(triangles),
			collisionMeshAsset.Translation,
			collisionMeshAsset.Rotation,
		))
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
