package game

import (
	"strings"
	"sync/atomic"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/log"
)

const (
	initialRevision int32 = -1
)

var (
	freeRevision int32
)

func nextRevision() int32 {
	return atomic.AddInt32(&freeRevision, 1)
}

// TransformFunc is a mechanism to calculate a custom absolute matrix
// off of the parent and the current matrix, circumventing standard
// matrix multiplication rules.
type TransformFunc func(parent, current dprec.Mat4) dprec.Mat4

// DefaultTransformFunc is a TransformFunc that applies standard matrix
// multiplication rules.
func DefaultTransformFunc(parent, current dprec.Mat4) dprec.Mat4 {
	return dprec.Mat4Prod(parent, current)
}

// NewNode creates a new detached Node instance.
func NewNode() *Node {
	return &Node{
		revision:  initialRevision,
		transform: DefaultTransformFunc,

		position: dprec.ZeroVec3(),
		rotation: dprec.IdentityQuat(),
		scale:    dprec.NewVec3(1.0, 1.0, 1.0),
	}
}

// Node represents a positioning of an object or part of one in the 3D scene.
type Node struct {
	name         string
	parent       *Node
	firstChild   *Node
	lastChild    *Node
	leftSibling  *Node
	rightSibling *Node

	position dprec.Vec3
	rotation dprec.Quat
	scale    dprec.Vec3

	// revision is a mechanism through which it is determined if the absolute
	// matrix cached for this Node is up to date. It borrows ideas from Lamport
	// timestamps used in distributed systems. The matrix is considered up to date
	// if the revision is larger than the parent's revision.
	revision  int32
	transform TransformFunc
	absMatrix dprec.Mat4

	body         *physics.Body
	mesh         *graphics.Mesh
	armature     *graphics.Armature
	armatureBone int
	camera       *graphics.Camera
	light        *graphics.Light
}

// PrintHierarchy prints debug information regarding the hierarchy that starts
// from this node.
func (n *Node) PrintHierarchy(depth int) {
	log.Info("%sNODE:%s", strings.Repeat(" ", depth), n.name)
	for child := n.FirstChild(); child != nil; child = child.RightSibling() {
		child.PrintHierarchy(depth + 2)
	}
}

// Name returns this Node's name.
func (n *Node) Name() string {
	return n.name
}

// SetName changes this Node's name.
func (n *Node) SetName(name string) {
	n.name = name
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
		translation, rotation, scale := absMatrix.TRS()
		n.SetPosition(translation)
		n.SetRotation(rotation)
		n.SetScale(scale)
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
	sibling.revision = initialRevision
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
	sibling.revision = initialRevision
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
	child.revision = initialRevision
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
	child.revision = initialRevision
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
	child.revision = initialRevision
}

// FindNode searches the hierarchy starting from this Node (inclusive) for a
// Node that has the specified name.
func (n *Node) FindNode(name string) *Node {
	if n.name == name {
		return n
	}
	for child := n.firstChild; child != nil; child = child.rightSibling {
		if result := child.FindNode(name); result != nil {
			return result
		}
	}
	return nil
}

// Position returns this Node's relative position to the parent.
func (n *Node) Position() dprec.Vec3 {
	return n.position
}

// SetPosition changes this Node's relative position to the parent.
func (n *Node) SetPosition(position dprec.Vec3) {
	if position != n.position {
		n.position = position
		n.revision = initialRevision
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
		n.revision = initialRevision
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
		n.revision = initialRevision
	}
}

// Matrix returns the matrix transformation of this Node relative to the
// parent.
func (n *Node) Matrix() dprec.Mat4 {
	return dprec.TRSMat4(n.position, n.rotation, n.scale)
}

// UseTransformation configures a TransformFunc to be used for calculating
// the absolute matrix of this node. Pass nil to restore default behavior.
func (n *Node) UseTransformation(transform TransformFunc) {
	if transform != nil {
		n.transform = transform
	} else {
		n.transform = DefaultTransformFunc
	}
}

// AbsoluteMatrix returns the matrix transformation of this Node relative
// to the root coordinate system.
func (n *Node) AbsoluteMatrix() dprec.Mat4 {
	if !n.isStaleHierarchy() {
		return n.absMatrix
	}
	if n.parent == nil {
		n.absMatrix = n.Matrix()
	} else {
		n.absMatrix = n.transform(n.parent.AbsoluteMatrix(), n.Matrix())
	}
	n.revision = nextRevision() // make sure to call this last
	return n.absMatrix
}

// Body returns the physics Body that is attached to this Node.
func (n *Node) Body() *physics.Body {
	return n.body
}

// SetBody attaches a physics Body to this node.
func (n *Node) SetBody(body *physics.Body) {
	n.body = body
}

// Mesh returns the graphics Mesh that is attached to this Node.
func (n *Node) Mesh() *graphics.Mesh {
	return n.mesh
}

// SetMesh attaches a graphics Mesh to this Node.
func (n *Node) SetMesh(mesh *graphics.Mesh) {
	n.mesh = mesh
}

// Camera returns the graphics Camera that is attached to this Node.
func (n *Node) Camera() *graphics.Camera {
	return n.camera
}

// SetCamera attaches a graphics Camera to this Node.
func (n *Node) SetCamera(camera *graphics.Camera) {
	n.camera = camera
}

// Armature returns the graphics Armature that is attached to this Node.
func (n *Node) Armature() *graphics.Armature {
	return n.armature
}

// SetArmature attaches a graphics Armature to this Node.
func (n *Node) SetArmature(armature *graphics.Armature) {
	n.armature = armature
}

// ArmatureBone returns the bone index if the Armature that is affected by
// this Node.
func (n *Node) ArmatureBone() int {
	return n.armatureBone
}

// SetArmatureBone configures the bone index of the Armature that is affected
// by this Node.
func (n *Node) SetArmatureBone(bone int) {
	n.armatureBone = bone
}

// Light returns the graphics Light that is attached to this Node.
func (n *Node) Light() *graphics.Light {
	return n.light
}

// SetLight attaches a graphics Light to this Node.
func (n *Node) SetLight(light *graphics.Light) {
	n.light = light
}

func (n *Node) isStaleHierarchy() bool {
	if n.revision == initialRevision {
		// Default revision is considered stale.
		return true
	}
	if n.parent == nil {
		// No parent and not at initial revision means it is not stale.
		return false
	}
	if n.revision <= n.parent.revision {
		// The parent is ahead so this node is stale.
		return true
	}
	// This node appears to be fine with relation to its parent but
	// it is unclear if the parent is up to date.
	return n.parent.isStaleHierarchy()
}
