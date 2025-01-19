package hierarchy

import "github.com/mokiat/gomath/dprec"

// TODO: Make it possible to have static nodes that have a fixed transform
// that does not change even in case of hierarchy changes.

// Node represents the transformation of an object or part of one in 3D space.
type Node struct {
	name string

	parent       *Node
	firstChild   *Node
	lastChild    *Node
	leftSibling  *Node
	rightSibling *Node

	position dprec.Vec3
	rotation dprec.Quat
	scale    dprec.Vec3

	transform TransformFunc
	source    NodeSource
	target    NodeTarget

	// revision is a mechanism through which it is determined if the absolute
	// matrix cached for this Node is up to date. It borrows ideas from Lamport
	// timestamps used in distributed systems. The matrix is considered up to date
	// if the revision is larger than the parent's revision.
	revision  int32
	absMatrix dprec.Mat4
}

// NewNode creates a new detached Node instance.
func NewNode() *Node {
	return &Node{
		revision: initialRevision,

		position: dprec.ZeroVec3(),
		rotation: dprec.IdentityQuat(),
		scale:    dprec.NewVec3(1.0, 1.0, 1.0),

		transform: DefaultTransformFunc,
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

// Detach removes this Node from the hierarchy.
func (n *Node) Detach() {
	if n.parent != nil {
		absMatrix := n.AbsoluteMatrix()
		n.parent.RemoveChild(n)
		translation, rotation, scale := absMatrix.TRS()
		// TODO: SetMatrix
		n.SetPosition(translation)
		n.SetRotation(rotation)
		n.SetScale(scale)
	}
}

// Delete removes this Node from the hierarchy and deletes all of its children.
//
// The node can be reused after deletion.
func (n *Node) Delete() {
	child := n.FirstChild()
	for child != nil {
		next := child.RightSibling()
		child.Delete()
		child = next
	}
	if n.source != nil {
		n.source.Release()
		n.source = nil
	}
	if n.target != nil {
		n.target.Release()
		n.target = nil
	}
	n.Detach()
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

// Visit traverses the node hierarchy starting from the current node.
func (n *Node) Visit(callback func(*Node)) {
	callback(n)
	for child := n.firstChild; child != nil; child = child.rightSibling {
		child.Visit(callback)
	}
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

// IsDescendantOf returns whether the node is a descendant of the specified
// ancestor.
func (n *Node) IsDescendantOf(ancestor *Node) bool {
	current := n
	for current != nil {
		if current == ancestor {
			return true
		}
		current = current.parent
	}
	return false
}

// UseTransformation configures a TransformFunc to be used for calculating
// the absolute matrix of this node. Pass nil to restore the default behavior.
func (n *Node) UseTransformation(transform TransformFunc) {
	if transform != nil {
		n.transform = transform
	} else {
		n.transform = DefaultTransformFunc
	}
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

// SetMatrix changes this node's relative transformation th the specified
// matrix.
func (n *Node) SetMatrix(matrix dprec.Mat4) {
	translation, rotation, scale := matrix.TRS()
	n.SetPosition(translation)
	n.SetRotation(rotation)
	n.SetScale(scale)
}

// BaseAbsoluteMatrix returns the absolute matrix of the parent, if there is
// one, otherwise the identity matrix.
func (n *Node) BaseAbsoluteMatrix() dprec.Mat4 {
	if n.parent == nil {
		return dprec.IdentityMat4()
	}
	return n.parent.AbsoluteMatrix()
}

// AbsoluteMatrix returns the matrix transformation of this Node relative
// to the root coordinate system.
func (n *Node) AbsoluteMatrix() dprec.Mat4 {
	if !n.isStaleHierarchy() {
		return n.absMatrix
	}
	n.absMatrix = n.transform(n)
	n.revision = nextRevision() // make sure to call this last
	return n.absMatrix
}

// SetAbsoluteMatrix changes the relative position, rotation and scale
// of this node based on the specified absolute transformation matrix.
func (n *Node) SetAbsoluteMatrix(matrix dprec.Mat4) {
	if n.parent == nil {
		n.SetMatrix(matrix)
	} else {
		parentMatrix := n.parent.AbsoluteMatrix()
		relativeMatrix := dprec.Mat4Prod(
			dprec.InverseMat4(parentMatrix),
			matrix,
		)
		n.SetMatrix(relativeMatrix)
	}
}

// Source returns the transformation input for this node.
func (n *Node) Source() NodeSource {
	return n.source
}

// SetSource changes the transformation input for this node.
func (n *Node) SetSource(source NodeSource) {
	n.source = source
}

// Target returns the transformation output for this node.
func (n *Node) Target() NodeTarget {
	return n.target
}

// SetTarget changes the transformation output for this node.
func (n *Node) SetTarget(target NodeTarget) {
	n.target = target
}

// ApplyFromSource requests that this node be updated based on its source.
// If recursive is specified, the same is applied down the hierarchy as well.
func (n *Node) ApplyFromSource(recursive bool) {
	if n.source != nil {
		n.source.ApplyTo(n)
	}
	if recursive {
		for child := n.firstChild; child != nil; child = child.rightSibling {
			child.ApplyFromSource(recursive)
		}
	}
}

// ApplyToTarget requests that this node be applied to its target.
// If recursive is specified, the same is applied down the hierarchy as well.
func (n *Node) ApplyToTarget(recursive bool) {
	if n.target != nil {
		n.target.ApplyFrom(n)
	}
	if recursive {
		for child := n.firstChild; child != nil; child = child.rightSibling {
			child.ApplyToTarget(recursive)
		}
	}
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
