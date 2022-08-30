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
	meshInstances   []MeshInstance
	bodyDefinitions []*physics.BodyDefinition
	bodyInstances   []bodyInstance

	// TODO: Fix these as well
	Animations []*AnimationDefinition
	Armatures  []*ArmatureDefinition
	Materials  []*MaterialDefinition
}

type MeshInstance struct {
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
			Position:    dprec.ArrayToVec3(nodeAsset.Translation),
			Rotation: dprec.NewQuat(
				nodeAsset.Rotation[3],
				nodeAsset.Rotation[0],
				nodeAsset.Rotation[1],
				nodeAsset.Rotation[2],
			),
			Scale: dprec.ArrayToVec3(nodeAsset.Scale),
		}
	}
	result.nodes = nodes

	bodyDefinitions := make([]*physics.BodyDefinition, len(modelAsset.BodyDefinitions))
	for i, definitionAsset := range modelAsset.BodyDefinitions {
		physicsEngine := r.engine.Physics()
		bodyDefinitions[i] = physicsEngine.CreateBodyDefinition(physics.BodyDefinitionInfo{
			Mass:                   definitionAsset.Mass,
			MomentOfInertia:        definitionAsset.MomentOfInertia,
			RestitutionCoefficient: definitionAsset.RestitutionCoefficient,
			DragFactor:             definitionAsset.DragFactor,
			AngularDragFactor:      definitionAsset.AngularDragFactor,
			CollisionShapes:        r.constructCollisionShapes(definitionAsset),
			AerodynamicShapes:      nil, // TODO
		})
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
