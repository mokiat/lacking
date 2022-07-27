package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
)

// NewNode creates a new detached Node instance.
func NewNode() *Node {
	return &Node{
		isStale:  true,
		position: dprec.ZeroVec3(),
		rotation: dprec.IdentityQuat(),
		scale:    dprec.NewVec3(1.0, 1.0, 1.0),
	}
}

// Node represents a positioning of an object or part of one in the 3D scene.
type Node struct {
	parent       *Node
	firstChild   *Node
	lastChild    *Node
	leftSibling  *Node
	rightSibling *Node

	isStale   bool
	position  dprec.Vec3
	rotation  dprec.Quat
	scale     dprec.Vec3
	absMatrix dprec.Mat4

	body   *physics.Body
	mesh   *graphics.Mesh
	camera *graphics.Camera
}

// Parent returns the parent Node in the hierarchy. If this is the top-most
// Node then nil is returned.
func (n *Node) Parent() *Node {
	return n.parent
}

// FirstChild returns the first (left-most) child Node of this Node. If this
// Node does not have any children then this method returns nil.
func (n *Node) FirstChild() *Node {
	return n.firstChild
}

// LastChild returns the last (right-most) child Node of this Node. If this
// Node does not have any children then this method returns nil.
func (n *Node) LastChild() *Node {
	return n.lastChild
}

// LeftSibling returns the left sibling Node of this Node. If this Node is the
// left-most child of its parent or does not have a parent then this method
// returns nil.
func (n *Node) LeftSibling() *Node {
	return n.leftSibling
}

// RightSibling returns the right sibling Node of this Node. If this Node is
// the right-most child of its parent or does not have a parent then this
// method returns nil.
func (n *Node) RightSibling() *Node {
	return n.rightSibling
}

// Detach removes this Node from the hierarchy but does not release any
// resources.
func (n *Node) Detach() {
	if n.parent != nil {
		absMatrix := n.AbsoluteMatrix()
		n.parent.RemoveChild(n)
		n.SetPosition(absMatrix.Translation())
		n.SetRotation(absMatrix.RotationQuat())
		n.SetScale(absMatrix.Scale())
	}
}

// PrependSibling attaches a Node to the left of the current one.
func (n *Node) PrependSibling(sibling *Node) {
	sibling.Detach()
	sibling.leftSibling = n.leftSibling
	if n.leftSibling != nil {
		n.leftSibling.rightSibling = sibling
	}
	sibling.rightSibling = n
	n.leftSibling = sibling
	if n.parent != nil && sibling.leftSibling == nil {
		n.parent.firstChild = sibling
	}
	sibling.isStale = true
}

// AppendSibling attaches a Node to the right of the current one.
func (n *Node) AppendSibling(sibling *Node) {
	sibling.Detach()
	sibling.rightSibling = n.rightSibling
	if n.rightSibling != nil {
		n.rightSibling.leftSibling = sibling
	}
	sibling.leftSibling = n
	n.rightSibling = sibling
	if n.parent != nil && sibling.rightSibling == nil {
		n.parent.lastChild = sibling
	}
	sibling.isStale = true
}

// PrependChild adds the specified Node as the left-most child of this Node.
// If the preprended Node already has a parent, it is first detached from that
// parent.
func (n *Node) PrependChild(child *Node) {
	child.Detach()
	child.parent = n
	child.leftSibling = nil
	child.rightSibling = n.firstChild
	if n.firstChild != nil {
		n.firstChild.leftSibling = child
	}
	n.firstChild = child
	if n.lastChild == nil {
		n.lastChild = child
	}
	child.isStale = true
}

// AppendChild adds the specified Node as the right-most child of this Node.
// If the appended Node already has a parent, it is first detached from that
// parent.
func (n *Node) AppendChild(child *Node) {
	child.Detach()
	child.parent = n
	child.leftSibling = n.lastChild
	child.rightSibling = nil
	if n.firstChild == nil {
		n.firstChild = child
	}
	if n.lastChild != nil {
		n.lastChild.rightSibling = child
	}
	n.lastChild = child
	child.isStale = true
}

// RemoveChild removes the specified Node from the list of children held by
// this Node. If the specified Node is not a child of this Node, then nothing
// happens.
func (n *Node) RemoveChild(child *Node) {
	if child.parent != n {
		return
	}
	if child.leftSibling != nil {
		child.leftSibling.rightSibling = child.rightSibling
	}
	if child.rightSibling != nil {
		child.rightSibling.leftSibling = child.leftSibling
	}
	if n.firstChild == child {
		n.firstChild = child.rightSibling
	}
	if n.lastChild == child {
		n.lastChild = child.leftSibling
	}
	child.parent = nil
	child.leftSibling = nil
	child.rightSibling = nil
	child.isStale = true
}

// Position returns this Node's relative position to the parent.
func (n *Node) Position() dprec.Vec3 {
	return n.position
}

// SetPosition changes this Node's relative position to the parent.
func (n *Node) SetPosition(position dprec.Vec3) {
	if position != n.position {
		n.position = position
		n.isStale = true
	}
}

// Rotation returns this Node's rotation relative to the parent.
func (n *Node) Rotation() dprec.Quat {
	return n.rotation
}

// SetRotation changes this Node's rotation relative to the parent.
func (n *Node) SetRotation(rotation dprec.Quat) {
	if rotation != n.rotation {
		n.rotation = dprec.UnitQuat(rotation)
		n.isStale = true
	}
}

// Scale returns this Node's scale relative to the parent.
func (n *Node) Scale() dprec.Vec3 {
	return n.scale
}

// SetScale changes this Node's scale relative to the parent.
func (n *Node) SetScale(scale dprec.Vec3) {
	if scale != n.scale {
		n.scale = scale
		n.isStale = true
	}
}

// Matrix returns the matrix transformation of this Node relative to the
// parent.
func (n *Node) Matrix() dprec.Mat4 {
	return dprec.TRSMat4(n.position, n.rotation, n.scale)
}

// AbsoluteMatrix returns the matrix transformation of this Node relative
// to the root coordinate system.
func (n *Node) AbsoluteMatrix() dprec.Mat4 {
	if n.parent == nil {
		if n.isStale {
			n.absMatrix = n.Matrix()
		}
	} else {
		if n.isStaleHierarchy() {
			n.absMatrix = dprec.Mat4Prod(n.parent.AbsoluteMatrix(), n.Matrix())
		}
	}
	n.isStale = false
	return n.absMatrix
}

func (n *Node) Body() *physics.Body {
	return n.body
}

func (n *Node) SetBody(body *physics.Body) {
	n.body = body
}

func (n *Node) Mesh() *graphics.Mesh {
	return n.mesh
}

func (n *Node) SetMesh(mesh *graphics.Mesh) {
	n.mesh = mesh
}

func (n *Node) Camera() *graphics.Camera {
	return n.camera
}

func (n *Node) SetCamera(camera *graphics.Camera) {
	n.camera = camera
}

func (n *Node) isStaleHierarchy() bool {
	if n.isStale {
		return true
	}
	if n.parent == nil {
		return false
	}
	return n.parent.isStaleHierarchy()
}
