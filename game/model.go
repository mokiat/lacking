package game

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
)

type NodeDefinition struct {
	Name        string
	Parent      *NodeDefinition
	Children    []*NodeDefinition
	Translation sprec.Vec3
	Rotation    sprec.Quat
	Scale       sprec.Vec3
}

type ArmatureDefinition struct {
	GraphicsTemplate *graphics.ArmatureTemplate
}

type MaterialDefinition struct {
}

type MeshInstanceDefinition struct {
	Name             string
	GraphicsTemplate *graphics.MeshTemplate
	Node             *NodeDefinition
	Armature         *ArmatureDefinition
}

type ModelDefinition struct {
	Nodes         []*NodeDefinition
	Armatures     []*ArmatureDefinition
	Materials     []*MaterialDefinition
	MeshInstances []*MeshInstanceDefinition
}

type Model struct {
	nodes     []*Node
	armatures []*graphics.Armature
	materials []*graphics.Material
}
