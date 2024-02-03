package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/hierarchy"
)

// NodeDefinition describes a node within a game scene.
type NodeDefinition struct {

	// Parent is the parent node of the node. If nil, the node is attached to
	// the root.
	Parent *hierarchy.Node

	// Name is the name of the node.
	Name string

	// Position is the relative position of the node.
	Position dprec.Vec3

	// Rotation is the relative rotation of the node.
	Rotation dprec.Quat

	// Scale is the relative scale of the node.
	Scale dprec.Vec3

	// IsStationary indicates whether the node is stationary.
	// Stationary nodes cannot be moved in the world, regardless of changes
	// to the parent or hierarchy.
	IsStationary bool

	// IsInseparable indicates whether the node is inseparable.
	// Inseparable nodes cannot be removed from their parents unless deleted.
	IsInseparable bool
}
