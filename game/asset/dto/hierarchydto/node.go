package hierarchydto

import "github.com/mokiat/gomath/dprec"

const (
	// UnspecifiedNodeID is the ID that is used to indicate that no node is
	// specified.
	UnspecifiedNodeID = uint32(0xFFFFFFFF)
)

// Node represents a single node in a model.
type Node struct {

	// ID is the unique identifier of the node within the file.
	ID uint32

	// ParentID is the ID of the parent node.
	//
	// If the node does not have a parent, this value is set to
	// UnspecifiedNodeID.
	ParentID uint32

	// Name is the name of the node.
	Name string

	// Translation is the translation of the node.
	Translation dprec.Vec3

	// Rotation is the rotation of the node.
	Rotation dprec.Quat

	// Scale is the scale of the node.
	Scale dprec.Vec3

	// Mask is the mask that specifies the behavior of the node.
	Mask NodeMask
}

// NodeMask specifies the behavior of a node.
type NodeMask uint32

const (
	// NodeMaskNone specifies that the node has no special behavior.
	NodeMaskNone NodeMask = 0

	// NodeMaskStationary specifies that the node is stationary and should not
	// be moved. The engine may optimize the node away.
	NodeMaskStationary NodeMask = 1 << iota

	// NodeMaskInseparable specifies that the node is inseparable from its
	// parent.
	NodeMaskInseparable
)
