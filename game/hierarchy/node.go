package hierarchy

import "github.com/mokiat/gomath/sprec"

func NewNode(parent *Node) *Node {
	return &Node{
		parent:         parent,
		parentRevision: -1,

		position: sprec.ZeroVec3(),
		rotation: sprec.IdentityQuat(),
		scale:    sprec.NewVec3(1.0, 1.0, 1.0),
		revision: 0,

		cachedRevision: -1,
	}
}

// Node represents a positioning of an object (model / camera) in the 3D scene.
type Node struct {
	parent         *Node
	parentRevision int

	position sprec.Vec3
	rotation sprec.Quat
	scale    sprec.Vec3
	revision int

	cachedMatrix   sprec.Mat4
	cachedRevision int
}

// Position returns this Node's relative position to the parent.
func (n *Node) Position() sprec.Vec3 {
	return n.position
}

// SetPosition changes this Node's relative position to the parent.
func (n *Node) SetPosition(position sprec.Vec3) {
	if position != n.position {
		n.position = position
		n.revision++
	}
}

// Rotation returns this Node's rotation relative to the parent.
func (n *Node) Rotation() sprec.Quat {
	return n.rotation
}

// SetRotation changes this Node's rotation relative to the parent.
func (n *Node) SetRotation(rotation sprec.Quat) {
	if rotation != n.rotation {
		n.rotation = sprec.UnitQuat(rotation)
		n.revision++
	}
}

// Scale returns this Node's scale relative to the parent.
func (n *Node) Scale() sprec.Vec3 {
	return n.scale
}

// SetScale changes this Node's scale relative to the parent.
func (n *Node) SetScale(scale sprec.Vec3) {
	if scale != n.scale {
		n.scale = scale
		n.revision++
	}
}

// Matrix returns the matrix transformation of this Node relative to the
// parent.
func (n *Node) Matrix() sprec.Mat4 {
	return sprec.TRSMat4(n.position, n.rotation, n.scale)
}

// AbsoluteMatrix returns the matrix transformation of this Node relative
// to the root coordinate system.
func (n *Node) AbsoluteMatrix() sprec.Mat4 {
	if n.parent == nil {
		if n.revision != n.cachedRevision {
			n.cachedMatrix = n.Matrix()
			n.cachedRevision = n.revision
		}
	} else {
		if n.parentRevision != n.parent.revision || n.revision != n.cachedRevision {
			n.parentRevision = n.parent.revision
			n.cachedMatrix = sprec.Mat4Prod(n.parent.AbsoluteMatrix(), n.Matrix())
			n.cachedRevision = n.revision
		}
	}
	return n.cachedMatrix
}
