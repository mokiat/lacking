package hierarchy

import (
	"iter"

	"github.com/mokiat/gomath/dprec"
)

// Node is a wrapper over a NodeID that simplifies interactions related to the
// node.
//
// It should not be persisted for long-term use.
type Node struct {
	scene *Scene
	id    NodeID
}

// ID returns the NodeID of the node.
func (n Node) ID() NodeID {
	return n.id
}

// IsNil returns whether the Node is nil.
func (n Node) IsNil() bool {
	return n.id.IsNil()
}

// IsValid returns whether the Node is valid.
func (n Node) IsValid() bool {
	return n.scene.IsValidNode(n.id)
}

// Delete removes the node from the scene.
func (n Node) Delete() {
	n.scene.DeleteNode(n.id)
}

// Name returns the name of the node.
func (n Node) Name() string {
	return n.scene.NodeName(n.id)
}

// SetName sets the name of the node.
func (n Node) SetName(name string) {
	n.scene.SetNodeName(n.id, name)
}

// IsVisible returns whether the node is visible.
func (n Node) IsVisible() bool {
	return n.scene.IsNodeVisible(n.id)
}

// SetVisible sets whether the node is visible.
func (n Node) SetVisible(visible bool) {
	n.scene.SetNodeVisible(n.id, visible)
}

// IsIndependent returns whether the node is independent (i.e., not affected by
// parent transforms).
func (n Node) IsIndependent() bool {
	return n.scene.IsNodeIndependent(n.id)
}

// SetIndependent sets whether the node is independent (i.e., not affected by
// parent transforms).
func (n Node) SetIndependent(independent bool) {
	n.scene.SetNodeIndependent(n.id, independent)
}

// TransformFunc returns the custom transformation function of the node.
func (n Node) TransformFunc() TransformFunc {
	return n.scene.NodeTransformFunc(n.id)
}

// SetTransformFunc sets the custom transformation function of the node.
func (n Node) SetTransformFunc(transformFunc TransformFunc) {
	n.scene.SetNodeTransformFunc(n.id, transformFunc)
}

// HasParent returns whether the node has a parent.
func (n Node) HasParent() bool {
	return n.scene.NodeHasParent(n.id)
}

// Parent returns the parent of the node, or a nil Node if it has no parent.
func (n Node) Parent() Node {
	return n.scene.Wrap(n.scene.NodeParent(n.id))
}

// FirstChild returns the first child of the node, or a nil Node if it has no
// children.
func (n Node) FirstChild() Node {
	return n.scene.Wrap(n.scene.NodeFirstChild(n.id))
}

// LastChild returns the last child of the node, or a nil Node if it has no
// children.
func (n Node) LastChild() Node {
	return n.scene.Wrap(n.scene.NodeLastChild(n.id))
}

// LeftSibling returns the left sibling of the node, or a nil Node if it has no
// left sibling.
func (n Node) LeftSibling() Node {
	return n.scene.Wrap(n.scene.NodeLeftSibling(n.id))
}

// RightSibling returns the right sibling of the node, or a nil Node if it has
// no right sibling.
func (n Node) RightSibling() Node {
	return n.scene.Wrap(n.scene.NodeRightSibling(n.id))
}

// ResetDelta resets the delta state of the node. If recursive is true, the
// delta state of all descendants is also reset.
func (n Node) ResetDelta(recursive bool) {
	n.scene.ResetNodeDelta(n.id, recursive)
}

// FindNode searches for a node with the specified name in the subtree rooted at
// this node. It returns a nil Node if no such node is found.
func (n Node) FindNode(name string) Node {
	return n.scene.Wrap(n.scene.FindSubtreeNode(n.id, name))
}

// IsAncestorOf returns whether the node is an ancestor of the specified
// descendant node.
func (n Node) IsAncestorOf(descendant Node) bool {
	return n.scene.IsNodeChain(n.id, descendant.id)
}

// PrependSibling attaches the specified sibling node as the left sibling of
// this node.
func (n Node) PrependSibling(sibling Node, preserveWorldTransform bool) {
	n.scene.PrependNodeSibling(n.id, sibling.id, preserveWorldTransform)
}

// AppendSibling attaches the specified sibling node as the right sibling of
// this node.
func (n Node) AppendSibling(sibling Node, preserveWorldTransform bool) {
	n.scene.AppendNodeSibling(n.id, sibling.id, preserveWorldTransform)
}

// PrependChild attaches the specified child node as the first child of this
// node.
func (n Node) PrependChild(child Node, preserveWorldTransform bool) {
	n.scene.PrependNodeChild(n.id, child.id, preserveWorldTransform)
}

// AppendChild attaches the specified child node as the last child of this
// node.
func (n Node) AppendChild(child Node, preserveWorldTransform bool) {
	n.scene.AppendNodeChild(n.id, child.id, preserveWorldTransform)
}

// Detach detaches the node from its parent. If preserveWorldTransform is true,
// the world transform of the node is preserved.
func (n Node) Detach(preserveWorldTransform bool) {
	n.scene.DetachNode(n.id, preserveWorldTransform)
}

// Position returns the local position of the node.
func (n Node) Position() dprec.Vec3 {
	return n.scene.NodePosition(n.id)
}

// SetPosition sets the local position of the node.
func (n Node) SetPosition(position dprec.Vec3) {
	n.scene.SetNodePosition(n.id, position)
}

// InitializePosition initializes the local position of the node with the
// specified ID, avoiding interpolation with previous value.
func (n Node) InitializePosition(position dprec.Vec3) {
	n.scene.InitializeNodePosition(n.id, position)
}

// Rotation returns the local rotation of the node.
func (n Node) Rotation() dprec.Quat {
	return n.scene.NodeRotation(n.id)
}

// SetRotation sets the local rotation of the node.
func (n Node) SetRotation(rotation dprec.Quat) {
	n.scene.SetNodeRotation(n.id, rotation)
}

// InitializeRotation initializes the local rotation of the node with the
// specified ID, avoiding interpolation with previous value.
func (n Node) InitializeRotation(rotation dprec.Quat) {
	n.scene.InitializeNodeRotation(n.id, rotation)
}

// Scale returns the local scale of the node.
func (n Node) Scale() dprec.Vec3 {
	return n.scene.NodeScale(n.id)
}

// SetScale sets the local scale of the node.
func (n Node) SetScale(scale dprec.Vec3) {
	n.scene.SetNodeScale(n.id, scale)
}

// InitializeScale initializes the local scale of the node with the specified
// ID, avoiding interpolation with previous value.
func (n Node) InitializeScale(scale dprec.Vec3) {
	n.scene.InitializeNodeScale(n.id, scale)
}

// Matrix returns the local transformation matrix of the node.
func (n Node) Matrix() dprec.Mat4 {
	return n.scene.NodeMatrix(n.id)
}

// SetMatrix sets the local transformation matrix of the node.
func (n Node) SetMatrix(matrix dprec.Mat4) {
	n.scene.SetNodeMatrix(n.id, matrix)
}

// InitializeMatrix initializes the local transformation matrix of the node
// with the specified ID, avoiding interpolation with previous value.
func (n Node) InitializeMatrix(matrix dprec.Mat4) {
	n.scene.InitializeNodeMatrix(n.id, matrix)
}

// BaseMatrix returns the base transformation matrix of the node with the
// specified ID - this is the absolute matrix of the parent, if there is one,
// or the identity matrix otherwise.
func (n Node) BaseMatrix() dprec.Mat4 {
	return n.scene.NodeBaseMatrix(n.id)
}

// AbsoluteMatrix returns the absolute transformation matrix of the node.
func (n Node) AbsoluteMatrix() dprec.Mat4 {
	return n.scene.NodeAbsoluteMatrix(n.id)
}

// SetAbsoluteMatrix sets the absolute transformation matrix of the node.
// This is a convenience method that sets the local transformation matrix
// accordingly.
func (n Node) SetAbsoluteMatrix(matrix dprec.Mat4) {
	n.scene.SetNodeAbsoluteMatrix(n.id, matrix)
}

// InterpolatedAbsoluteMatrix returns the interpolated absolute transformation
// matrix of the node, based on the specified alpha value in the range [0, 1].
func (n Node) InterpolatedAbsoluteMatrix(alpha float64) dprec.Mat4 {
	return n.scene.NodeInterpolatedAbsoluteMatrix(n.id, alpha)
}

// Visit visits the node and all its descendants in depth-first order, calling
// the specified visitor function for each node. If the visitor function returns
// false, the traversal is stopped.
func (n Node) Visit(visitor func(Node) bool) {
	n.scene.VisitSubtree(n.id, func(id NodeID) bool {
		return visitor(n.scene.Wrap(id))
	})
}

// TreeIter returns an iterator that traverses the node and all its descendants
// in depth-first order.
func (n Node) TreeIter() iter.Seq[Node] {
	return func(yield func(Node) bool) {
		for id := range n.scene.SubtreeIter(n.id) {
			if !yield(n.scene.Wrap(id)) {
				return
			}
		}
	}
}
