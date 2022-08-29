package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
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
	nodes      []NodeDefinition
	Animations []*AnimationDefinition
	Armatures  []*ArmatureDefinition
	Materials  []*MaterialDefinition

	meshDefinitions []*graphics.MeshDefinition
	meshInstances   []MeshInstance

	bodyDefinitions []*physics.BodyDefinition
	bodyInstances   []BodyInstance
}

type MeshInstance struct {
	Name            string
	NodeIndex       int
	DefinitionIndex int
	// Armature        *ArmatureDefinition
}

type BodyInstance struct {
	Name            string
	NodeIndex       int
	DefinitionIndex int
	Position        dprec.Vec3
	Rotation        dprec.Quat
	IsDynamic       bool
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
