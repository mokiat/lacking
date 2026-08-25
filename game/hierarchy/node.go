package hierarchy

import (
	"iter"

	"github.com/mokiat/gomath/dprec"
)

// NodeID uniquely identifies a node within a [Scene].
//
// A NodeID remains stable across the lifetime of the node it references but is
// invalidated once that node is deleted, even if the node's storage slot is
// later reused by a different node. The zero value equals [NilNodeID].
type NodeID struct {
	index    int32
	revision int32
}

// NilNodeID is a sentinel value representing a nil NodeID.
var NilNodeID = NodeID{}

// NodeView provides access to the nodes of a [Scene].
//
// It exposes operations to create, delete, traverse, and modify nodes. A
// NodeView is a lightweight value that can be freely copied and passed around;
// all instances obtained from the same [Scene] observe the same underlying
// state.
type NodeView struct {
	scene *Scene
}

// Create creates a new node and returns its [NodeID].
//
// The new node is created as a root node with no children, an empty name, and
// an identity transformation.
func (v NodeView) Create() NodeID {
	index, node := v.scene.allocateNode()

	*node = nodeState{
		parentIndex:      nilIndex,
		nextSiblingIndex: nilIndex,
		firstChildIndex:  nilIndex,
		revision:         node.revision + 1, // progress revision to valid (odd) value
		flags:            0,
		name:             "",
		position:         dprec.ZeroVec3(),
		rotation:         dprec.IdentityQuat(),
		scale:            dprec.NewVec3(1.0, 1.0, 1.0),
		absMatrix:        dprec.IdentityMat4(),
		oldAbsMatrix:     dprec.IdentityMat4(),
	}

	return NodeID{
		index:    index,
		revision: node.revision,
	}
}

// CreateHandle creates a new node and returns a [NodeHandle] to it.
//
// It is a convenience equivalent to passing the result of [NodeView.Create] to
// [NodeView.Handle].
func (v NodeView) CreateHandle() NodeHandle {
	return v.Handle(v.Create())
}

// Delete removes the node with the specified ID from the scene.
//
// All descendants of the node are deleted as well. Any [NodeID] referring to a
// deleted node becomes invalid.
func (v NodeView) Delete(id NodeID) {
	index, node := v.resolve(id, true)

	for _, binding := range v.scene.bindings {
		binding.handleNodeDelete(id)
	}

	v.Detach(id, false)

	for node.firstChildIndex != nilIndex {
		v.Delete(v.idFromIndex(node.firstChildIndex))
	}

	*node = nodeState{
		parentIndex:      nilIndex,
		nextSiblingIndex: nilIndex,
		firstChildIndex:  nilIndex,
		revision:         node.revision + 1, // progress revision to invalid (even) value
		flags:            0,
		name:             "",
		position:         dprec.ZeroVec3(),
		rotation:         dprec.IdentityQuat(),
		scale:            dprec.NewVec3(1.0, 1.0, 1.0),
		absMatrix:        dprec.IdentityMat4(),
		oldAbsMatrix:     dprec.IdentityMat4(),
	}

	v.scene.releaseNode(index)
}

// Each invokes the specified callback for each valid node in the scene.
//
// The callback should return true to continue the iteration or false to stop
// it early. The order in which nodes are visited is unspecified.
func (v NodeView) Each(cb func(NodeID) bool) {
	for i := range v.scene.nodes {
		node := &v.scene.nodes[i]
		if !node.isValid() {
			continue
		}
		id := NodeID{
			index:    int32(i),
			revision: node.revision,
		}
		if !cb(id) {
			return
		}
	}
}

// Iter returns an iterator over all valid nodes in the scene.
//
// The order in which nodes are visited is unspecified.
func (v NodeView) Iter() iter.Seq[NodeID] {
	return func(yield func(NodeID) bool) {
		v.Each(yield)
	}
}

// EachRoot invokes the specified callback for each valid root node in the scene.
//
// A root node is one without a parent. The callback should return true to
// continue the iteration or false to stop it early. The order in which nodes are
// visited is unspecified.
func (v NodeView) EachRoot(cb func(NodeID) bool) {
	for i := range v.scene.nodes {
		node := &v.scene.nodes[i]
		if !node.isValid() || node.parentIndex != nilIndex {
			continue
		}
		id := NodeID{
			index:    int32(i),
			revision: node.revision,
		}
		if !cb(id) {
			return
		}
	}
}

// RootIter returns an iterator over all valid root nodes in the scene.
//
// A root node is one without a parent. The order in which nodes are visited is
// unspecified.
func (v NodeView) RootIter() iter.Seq[NodeID] {
	return func(yield func(NodeID) bool) {
		v.EachRoot(yield)
	}
}

// Walk traverses the whole scene hierarchy, invoking the specified callback for
// each valid node.
//
// Each node is visited before its children (depth-first, pre-order). The order
// in which the separate root subtrees are visited relative to one another is
// unspecified. The callback should return true to continue the traversal or
// false to stop it early.
func (v NodeView) Walk(cb func(NodeID) bool) {
	v.EachRoot(func(rootID NodeID) bool {
		return v.yieldSubtree(rootID.index, cb)
	})
}

// WalkIter returns an iterator that traverses the whole scene hierarchy
// depth-first, visiting each node before its children.
//
// See [NodeView.Walk] for the traversal order.
func (v NodeView) WalkIter() iter.Seq[NodeID] {
	return func(yield func(NodeID) bool) {
		v.Walk(yield)
	}
}

// WalkSubtree traverses the subtree rooted at the node with the specified ID,
// invoking the specified callback for each node.
//
// The traversal includes the root node itself as well as all of its
// descendants, visiting each node before its children (depth-first, pre-order).
// The callback should return true to continue the traversal or false to stop it
// early. If rootID does not refer to a valid node, the traversal is a no-op.
func (v NodeView) WalkSubtree(rootID NodeID, cb func(NodeID) bool) {
	if v.IsValid(rootID) {
		v.yieldSubtree(rootID.index, cb)
	}
}

// WalkSubtreeIter returns an iterator that traverses the subtree rooted at the
// node with the specified ID, visiting each node before its children.
//
// See [NodeView.WalkSubtree] for the traversal order and the handling of an
// invalid root.
func (v NodeView) WalkSubtreeIter(rootID NodeID) iter.Seq[NodeID] {
	return func(yield func(NodeID) bool) {
		v.WalkSubtree(rootID, yield)
	}
}

// SubtreeContains returns whether the node with the specified findID lies
// within the subtree rooted at the node with the specified rootID.
//
// The subtree includes the root node itself, so a node contains itself. It
// returns false if either rootID or findID does not refer to a valid node.
func (v NodeView) SubtreeContains(rootID, findID NodeID) bool {
	if !v.IsValid(rootID) || !v.IsValid(findID) {
		return false
	}
	found := false
	v.WalkSubtree(rootID, func(id NodeID) bool {
		if id == findID {
			found = true
			return false
		}
		return true
	})
	return found
}

// FindNode returns the ID of a node with the specified name, or [NilNodeID] if
// no such node exists.
//
// If multiple nodes share the name, which one is returned is unspecified, so
// this is best used with names that are unique within the scene. The name is
// matched exactly.
func (v NodeView) FindNode(name string) NodeID {
	for i := range v.scene.nodes {
		node := &v.scene.nodes[i]
		if !node.isValid() {
			continue
		}
		if node.name == name {
			return v.idFromIndex(int32(i))
		}
	}
	return NilNodeID
}

// FindNodeInSubtree returns the ID of a node with the specified name within the
// subtree rooted at the node with the specified ID, or [NilNodeID] if no such
// node exists.
//
// The search includes the root node itself as well as all of its descendants.
// If rootID is [NilNodeID], the entire scene is searched, equivalent to
// [NodeView.FindNode]. If multiple nodes in the subtree share the name, which
// one is returned is unspecified. The name is matched exactly.
func (v NodeView) FindNodeInSubtree(rootID NodeID, name string) NodeID {
	if rootID == NilNodeID {
		return v.FindNode(name)
	}

	if !v.IsValid(rootID) {
		return NilNodeID
	}

	index := v.findNodeInSubtree(rootID.index, name)
	if index == nilIndex {
		return NilNodeID

	}
	return v.idFromIndex(index)
}

// Handle returns a [NodeHandle] that wraps the specified node ID.
func (v NodeView) Handle(id NodeID) NodeHandle {
	return NodeHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID refers to a node that currently
// exists in the scene.
//
// It returns false for [NilNodeID] and for IDs whose node has been deleted.
func (v NodeView) IsValid(id NodeID) bool {
	index, _ := v.resolve(id, false)
	return index != nilIndex
}

// IsRoot returns whether the specified node has no parent.
func (v NodeView) IsRoot(id NodeID) bool {
	_, node := v.resolve(id, true)
	return node.parentIndex == nilIndex
}

// Parent returns the ID of the parent of the specified node, or [NilNodeID] if
// the node is a root.
func (v NodeView) Parent(id NodeID) NodeID {
	_, node := v.resolve(id, true)
	if node.parentIndex == nilIndex {
		return NilNodeID
	}
	return v.idFromIndex(node.parentIndex)
}

// NextSibling returns the ID of the next sibling of the specified node, or
// [NilNodeID] if the node has no further siblings.
//
// Iterating from [NodeView.FirstChild] through NextSibling visits all children
// of a node.
func (v NodeView) NextSibling(id NodeID) NodeID {
	_, node := v.resolve(id, true)
	if node.nextSiblingIndex == nilIndex {
		return NilNodeID
	}
	return v.idFromIndex(node.nextSiblingIndex)
}

// FirstChild returns the ID of the first child of the specified node, or
// [NilNodeID] if the node has no children.
func (v NodeView) FirstChild(id NodeID) NodeID {
	_, node := v.resolve(id, true)
	if node.firstChildIndex == nilIndex {
		return NilNodeID
	}
	return v.idFromIndex(node.firstChildIndex)
}

// Detach detaches the specified node from its parent, turning it into a root
// node. It is a no-op if the node is already a root.
//
// If preserveWorldTransform is true, the node's local transformation is
// adjusted so that its position, rotation, and scale in world space are
// unchanged by the detachment.
func (v NodeView) Detach(id NodeID, preserveWorldTransform bool) {
	index, node := v.resolve(id, true)

	if node.parentIndex == nilIndex {
		return // no-op
	}

	if preserveWorldTransform {
		v.refreshAbsoluteMatrix(node) // needed for reconstruction
	}

	v.detachNode(index)

	if preserveWorldTransform {
		v.reconstructLocalTransform(node) // restore local transform
	} else {
		v.markDirty(node)
	}

	v.refreshAbsoluteHidden(node, false)
}

// AttachChild attaches the node with the specified child ID as a child of the
// node with the specified parent ID.
//
// If the child already has a parent, it is first detached from it. Attaching a
// child to its current parent is a no-op.
//
// If preserveWorldTransform is true, the child's local transformation is
// adjusted so that its position, rotation, and scale in world space are
// unchanged by the reparenting.
func (v NodeView) AttachChild(parentID, childID NodeID, preserveWorldTransform bool) {
	parentIndex, parent := v.resolve(parentID, true)
	childIndex, child := v.resolve(childID, true)

	if child.parentIndex == parentIndex {
		return // no-op
	}

	if preserveWorldTransform {
		v.refreshAbsoluteMatrix(child)  // needed for reconstruction
		v.refreshAbsoluteMatrix(parent) // needed for reconstruction
	}

	if child.parentIndex != nilIndex {
		v.detachNode(childIndex)
	}
	v.attachChildNode(parentIndex, childIndex)

	if preserveWorldTransform {
		v.reconstructLocalTransform(child)
	} else {
		v.markDirty(child)
	}

	v.refreshAbsoluteHidden(child, parent.isAbsoluteHidden())
}

// IsIndependent returns whether the specified node's transformation is
// independent of its parent.
//
// See [NodeView.SetIndependent] for the meaning of independence. Nodes are not
// independent by default.
func (v NodeView) IsIndependent(id NodeID) bool {
	_, node := v.resolve(id, true)
	return node.hasFlag(nodeFlagIndependent)
}

// SetIndependent sets whether the specified node's transformation is independent
// of its parent.
//
// An independent node keeps its place in the hierarchy - it still has a parent,
// participates in traversal, and inherits hidden state - but its local
// transformation is used directly as its absolute transformation, ignoring the
// transformations of its ancestors, as if it were a root. A root node is always
// effectively independent.
//
// If preserveWorldTransform is true, the node's local transformation is adjusted
// so that its absolute (world) transformation is unchanged by the switch.
func (v NodeView) SetIndependent(id NodeID, independent, preserveWorldTransform bool) {
	_, node := v.resolve(id, true)

	if independent != node.hasFlag(nodeFlagIndependent) {
		v.changeNodeIndependent(node, independent, preserveWorldTransform)
	}
}

// IsHidden returns whether the specified node is explicitly hidden.
//
// This reflects only the node's own hidden state, as set via
// [NodeView.SetHidden], and is independent of whether any ancestor is hidden.
// See [NodeView.IsAbsoluteHidden] for the effective visibility.
func (v NodeView) IsHidden(id NodeID) bool {
	_, node := v.resolve(id, true)
	return node.hasFlag(nodeFlagHidden)
}

// IsVisible returns whether the specified node is not explicitly hidden.
//
// It is the negation of [NodeView.IsHidden] and, like it, reflects only the
// node's own hidden state, independent of whether any ancestor is hidden. See
// [NodeView.IsAbsoluteVisible] for the effective visibility.
func (v NodeView) IsVisible(id NodeID) bool {
	return !v.IsHidden(id)
}

// IsAbsoluteHidden returns whether the specified node is effectively hidden,
// either because it is hidden itself or because one of its ancestors is hidden.
func (v NodeView) IsAbsoluteHidden(id NodeID) bool {
	_, node := v.resolve(id, true)
	return node.isAbsoluteHidden()
}

// IsAbsoluteVisible returns whether the specified node is effectively visible,
// meaning that neither it nor any of its ancestors is hidden.
//
// It is the negation of [NodeView.IsAbsoluteHidden].
func (v NodeView) IsAbsoluteVisible(id NodeID) bool {
	return !v.IsAbsoluteHidden(id)
}

// SetHidden sets whether the specified node is hidden.
//
// Hiding a node causes it and all of its descendants to become absolutely
// hidden (see [NodeView.IsAbsoluteHidden]). A descendant remains absolutely
// hidden until every ancestor that hides it, as well as the descendant itself,
// is no longer hidden.
func (v NodeView) SetHidden(id NodeID, hidden bool) {
	_, node := v.resolve(id, true)
	if hidden != node.hasFlag(nodeFlagHidden) {
		v.changeNodeHidden(node, hidden)
	}
}

// SetVisible sets whether the specified node is visible.
//
// It is the inverse of [NodeView.SetHidden]; making a node invisible hides it,
// causing it and all of its descendants to become absolutely hidden (see
// [NodeView.IsAbsoluteHidden]).
func (v NodeView) SetVisible(id NodeID, visible bool) {
	v.SetHidden(id, !visible)
}

// Name returns the name of the specified node.
func (v NodeView) Name(id NodeID) string {
	_, node := v.resolve(id, true)
	return node.name
}

// SetName sets the name of the specified node.
//
// Names need not be unique.
func (v NodeView) SetName(id NodeID, name string) {
	_, node := v.resolve(id, true)
	node.name = name
}

// Snap records the current absolute transformations of the specified node and
// all of its descendants as their previous transformations, collapsing any
// pending interpolation so that [NodeView.InterpolatedAbsoluteMatrix] returns
// the current pose for every fraction.
//
// Use this to teleport a node: after moving it (and hence its descendants) by a
// discontinuous amount, snapping prevents rendering from smoothly interpolating
// across the gap, which would otherwise make the node appear to slide from its
// old position to the new one. Snap after applying the new transformation, so
// that the current pose is the one recorded.
//
// This is not part of the per-frame update loop; for that, the scene as a whole
// is advanced once per fixed-rate step via [Scene.AdvanceStep].
func (v NodeView) Snap(id NodeID) {
	_, node := v.resolve(id, true)
	v.snapNode(node, true)
}

// Position returns the local position of the specified node, relative to its
// parent.
func (v NodeView) Position(id NodeID) dprec.Vec3 {
	_, node := v.resolve(id, true)
	return node.position
}

// SetPosition sets the local position of the specified node, relative to its
// parent.
func (v NodeView) SetPosition(id NodeID, position dprec.Vec3) {
	_, node := v.resolve(id, true)
	if position != node.position {
		node.position = position
		v.markDirty(node)
	}
}

// Rotation returns the local rotation of the specified node, relative to its
// parent.
func (v NodeView) Rotation(id NodeID) dprec.Quat {
	_, node := v.resolve(id, true)
	return node.rotation
}

// SetRotation sets the local rotation of the specified node, relative to its
// parent.
func (v NodeView) SetRotation(id NodeID, rotation dprec.Quat) {
	_, node := v.resolve(id, true)
	if rotation != node.rotation {
		node.rotation = rotation
		v.markDirty(node)
	}
}

// Scale returns the local scale of the specified node, relative to its parent.
func (v NodeView) Scale(id NodeID) dprec.Vec3 {
	_, node := v.resolve(id, true)
	return node.scale
}

// SetScale sets the local scale of the specified node, relative to its parent.
func (v NodeView) SetScale(id NodeID, scale dprec.Vec3) {
	_, node := v.resolve(id, true)
	if scale != node.scale {
		node.scale = scale
		v.markDirty(node)
	}
}

// TRS returns the local translation, rotation, and scale of the specified node,
// relative to its parent.
func (v NodeView) TRS(id NodeID) (dprec.Vec3, dprec.Quat, dprec.Vec3) {
	_, node := v.resolve(id, true)
	return node.position, node.rotation, node.scale
}

// SetTRS sets the local translation, rotation, and scale of the specified node,
// relative to its parent, in a single operation.
func (v NodeView) SetTRS(id NodeID, position dprec.Vec3, rotation dprec.Quat, scale dprec.Vec3) {
	_, node := v.resolve(id, true)
	if (position != node.position) || (rotation != node.rotation) || (scale != node.scale) {
		node.position = position
		node.rotation = rotation
		node.scale = scale
		v.markDirty(node)
	}
}

// Matrix returns the local transformation matrix of the specified node,
// relative to its parent.
//
// It is composed from the node's position, rotation, and scale.
func (v NodeView) Matrix(id NodeID) dprec.Mat4 {
	position, rotation, scale := v.TRS(id)
	return dprec.TRSMat4(position, rotation, scale)
}

// SetMatrix sets the local transformation of the specified node from the
// specified matrix, relative to its parent.
//
// The matrix is decomposed into position, rotation, and scale.
func (v NodeView) SetMatrix(id NodeID, matrix dprec.Mat4) {
	position, rotation, scale := matrix.TRS()
	v.SetTRS(id, position, rotation, scale)
}

// ReferenceMatrix returns the absolute transformation matrix that serves as the
// reference frame for the specified node's local transformation.
//
// For a node with a parent this is the parent's absolute matrix; for a root or
// independent node (see [NodeView.IsIndependent]) it is the identity matrix.
// The node's absolute matrix equals its reference matrix composed with its
// local matrix.
func (v NodeView) ReferenceMatrix(id NodeID) dprec.Mat4 {
	_, node := v.resolve(id, true)
	if node.isTransformIndependent() {
		return dprec.IdentityMat4()
	}
	parent := &v.scene.nodes[node.parentIndex]
	v.refreshAbsoluteMatrix(parent)
	return parent.absMatrix
}

// AbsoluteMatrix returns the absolute (world) transformation matrix of the
// specified node.
//
// It is the node's local transformation composed with the transformations of
// all its ancestors. For a root or independent node (see
// [NodeView.IsIndependent]) it equals the node's local matrix.
func (v NodeView) AbsoluteMatrix(id NodeID) dprec.Mat4 {
	_, node := v.resolve(id, true)
	v.refreshAbsoluteMatrix(node)
	return node.absMatrix
}

// SetAbsoluteMatrix sets the local transformation of the specified node so that
// its absolute (world) matrix equals the specified matrix.
//
// The local transformation is derived relative to the node's parent, so that
// changing an ancestor afterwards moves the node with it. For a root or
// independent node (see [NodeView.IsIndependent]) the local transformation is
// set directly. The absolute matrices of any descendants are updated to account
// for the change.
func (v NodeView) SetAbsoluteMatrix(id NodeID, matrix dprec.Mat4) {
	_, node := v.resolve(id, true)
	v.refreshAbsoluteMatrix(node)

	if node.absMatrix == matrix {
		return // no-op
	}

	if node.isTransformIndependent() {
		position, rotation, scale := matrix.TRS()
		node.position = position
		node.rotation = rotation
		node.scale = scale
		node.absMatrix = node.calculateMatrix()
		v.markChildrenDirty(node)
		return
	}

	parent := &v.scene.nodes[node.parentIndex]
	v.refreshAbsoluteMatrix(parent)

	relativeMatrix := dprec.Mat4Prod(
		dprec.InverseMat4(parent.absMatrix),
		matrix,
	)

	position, rotation, scale := relativeMatrix.TRS()
	node.position = position
	node.rotation = rotation
	node.scale = scale
	node.absMatrix = matrix

	v.markChildrenDirty(node)
}

// InterpolatedAbsoluteMatrix returns the specified node's absolute (world)
// transformation matrix interpolated between its previous and current values by
// the specified fraction.
//
// A fraction of 0 yields the previous absolute matrix and a fraction of 1
// yields the current one. The translation and scale are interpolated linearly
// and the rotation spherically. This is intended for smooth rendering between
// fixed-rate updates.
func (v NodeView) InterpolatedAbsoluteMatrix(id NodeID, fraction float64) dprec.Mat4 {
	_, node := v.resolve(id, true)

	v.refreshAbsoluteMatrix(node)
	if node.absMatrix == node.oldAbsMatrix {
		return node.absMatrix
	}

	oldPosition, oldRotation, oldScale := node.oldAbsMatrix.TRS()
	newPosition, newRotation, newScale := node.absMatrix.TRS()

	return dprec.TRSMat4(
		dprec.Vec3Lerp(oldPosition, newPosition, fraction),
		dprec.QuatSlerp(oldRotation, newRotation, fraction),
		dprec.Vec3Lerp(oldScale, newScale, fraction),
	)
}

func (v NodeView) idFromIndex(index int32) NodeID {
	node := &v.scene.nodes[index]
	return NodeID{
		index:    index,
		revision: node.revision,
	}
}

func (v NodeView) resolve(id NodeID, required bool) (int32, *nodeState) {
	if id.revision == 0 {
		if required {
			panic("invalid node ID")
		}
		return nilIndex, nil
	}
	node := &v.scene.nodes[id.index]
	if node.revision != id.revision {
		if required {
			panic("invalid node ID")
		}
		return nilIndex, nil
	}
	return id.index, node
}

func (v NodeView) attachChildNode(parentIndex, childIndex int32) {
	parent := &v.scene.nodes[parentIndex]
	child := &v.scene.nodes[childIndex]

	child.parentIndex = parentIndex
	child.nextSiblingIndex = parent.firstChildIndex
	parent.firstChildIndex = childIndex
}

func (v NodeView) detachNode(index int32) {
	node := &v.scene.nodes[index]

	parent := &v.scene.nodes[node.parentIndex]
	if parent.firstChildIndex == index {
		parent.firstChildIndex = node.nextSiblingIndex
	} else {
		prevIndex := parent.firstChildIndex
		for prevIndex != nilIndex {
			child := &v.scene.nodes[prevIndex]
			if child.nextSiblingIndex == index {
				child.nextSiblingIndex = node.nextSiblingIndex
				break
			}
			prevIndex = child.nextSiblingIndex
		}
	}

	node.parentIndex = nilIndex
	node.nextSiblingIndex = nilIndex
}

func (v NodeView) yieldSubtree(index int32, cb func(NodeID) bool) bool {
	if !cb(v.idFromIndex(index)) {
		return false
	}
	node := &v.scene.nodes[index]
	childIndex := node.firstChildIndex
	for childIndex != nilIndex {
		child := &v.scene.nodes[childIndex]
		if !v.yieldSubtree(childIndex, cb) {
			return false
		}
		childIndex = child.nextSiblingIndex
	}
	return true
}

func (v NodeView) findNodeInSubtree(index int32, name string) int32 {
	node := &v.scene.nodes[index]
	if node.name == name {
		return index
	}

	childIndex := node.firstChildIndex
	for childIndex != nilIndex {
		foundIndex := v.findNodeInSubtree(childIndex, name)
		if foundIndex != nilIndex {
			return foundIndex
		}
		child := &v.scene.nodes[childIndex]
		childIndex = child.nextSiblingIndex
	}

	return nilIndex
}

func (v NodeView) changeNodeHidden(node *nodeState, hidden bool) {
	oldAbsoluteHidden := node.isAbsoluteHidden()
	if hidden {
		node.setFlag(nodeFlagHidden)
	} else {
		node.unsetFlag(nodeFlagHidden)
	}
	newAbsoluteHidden := node.isAbsoluteHidden()

	if newAbsoluteHidden != oldAbsoluteHidden {
		childIndex := node.firstChildIndex
		for childIndex != nilIndex {
			child := &v.scene.nodes[childIndex]
			v.refreshAbsoluteHidden(child, newAbsoluteHidden)
			childIndex = child.nextSiblingIndex
		}
	}
}

func (v NodeView) changeNodeParentHidden(node *nodeState, parentHidden bool) {
	oldAbsoluteHidden := node.isAbsoluteHidden()
	if parentHidden {
		node.setFlag(nodeFlagParentHidden)
	} else {
		node.unsetFlag(nodeFlagParentHidden)
	}
	newAbsoluteHidden := node.isAbsoluteHidden()

	if newAbsoluteHidden != oldAbsoluteHidden {
		childIndex := node.firstChildIndex
		for childIndex != nilIndex {
			child := &v.scene.nodes[childIndex]
			v.refreshAbsoluteHidden(child, node.isAbsoluteHidden())
			childIndex = child.nextSiblingIndex
		}
	}
}

func (v NodeView) refreshAbsoluteHidden(node *nodeState, parentHidden bool) {
	if parentHidden != node.hasFlag(nodeFlagParentHidden) {
		v.changeNodeParentHidden(node, parentHidden)
	}
}

func (v NodeView) markDirty(node *nodeState) {
	if !node.hasFlag(nodeFlagDirty) {
		node.setFlag(nodeFlagDirty)
		v.markChildrenDirty(node)
	}
}

func (v NodeView) markChildrenDirty(node *nodeState) {
	childIndex := node.firstChildIndex
	for childIndex != nilIndex {
		child := &v.scene.nodes[childIndex]
		v.markDirty(child)
		childIndex = child.nextSiblingIndex
	}
}

func (v NodeView) changeNodeIndependent(node *nodeState, independent, preserveWorldTransform bool) {
	if preserveWorldTransform {
		v.refreshAbsoluteMatrix(node) // needed for reconstruction
	}

	if independent {
		node.setFlag(nodeFlagIndependent)
	} else {
		node.unsetFlag(nodeFlagIndependent)
	}

	if preserveWorldTransform {
		if !node.isTransformIndependent() {
			parent := &v.scene.nodes[node.parentIndex]
			v.refreshAbsoluteMatrix(parent) // needed for reconstruction
		}
		v.reconstructLocalTransform(node) // restore local transform
	} else {
		v.markDirty(node)
	}
}

func (v NodeView) snapNode(node *nodeState, recursive bool) {
	v.refreshAbsoluteMatrix(node)
	node.oldAbsMatrix = node.absMatrix

	if recursive {
		childIndex := node.firstChildIndex
		for childIndex != nilIndex {
			child := &v.scene.nodes[childIndex]
			v.snapNode(child, recursive)
			childIndex = child.nextSiblingIndex
		}
	}
}

func (v NodeView) refreshAbsoluteMatrix(node *nodeState) {
	if !node.hasFlag(nodeFlagDirty) {
		return
	}
	node.unsetFlag(nodeFlagDirty)

	if node.isTransformIndependent() {
		node.absMatrix = node.calculateMatrix()
		return
	}

	parent := &v.scene.nodes[node.parentIndex]
	v.refreshAbsoluteMatrix(parent)

	node.absMatrix = dprec.Mat4Prod(
		parent.absMatrix,
		node.calculateMatrix(),
	)
}

func (v NodeView) reconstructLocalTransform(node *nodeState) {
	if node.isTransformIndependent() {
		position, rotation, scale := node.absMatrix.TRS()
		node.position = position
		node.rotation = rotation
		node.scale = scale
		return
	}

	parent := &v.scene.nodes[node.parentIndex]
	relativeMatrix := dprec.Mat4Prod(
		dprec.InverseMat4(parent.absMatrix),
		node.absMatrix,
	)

	position, rotation, scale := relativeMatrix.TRS()
	node.position = position
	node.rotation = rotation
	node.scale = scale
}

// NodeHandle is a convenience wrapper that binds a [NodeID] to the [NodeView]
// it belongs to, allowing a node to be operated on without passing the view and
// ID separately.
//
// A handle to [NilNodeID] or to a deleted node is not valid; see
// [NodeHandle.IsValid].
type NodeHandle struct {
	view NodeView
	id   NodeID
}

// ID returns the ID of the node wrapped by this handle.
func (h NodeHandle) ID() NodeID {
	return h.id
}

// Delete removes the wrapped node from the scene.
//
// It is equivalent to [NodeView.Delete].
func (h NodeHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether the wrapped node currently exists in the scene.
//
// It is equivalent to [NodeView.IsValid].
func (h NodeHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// IsRoot returns whether the wrapped node has no parent.
//
// It is equivalent to [NodeView.IsRoot].
func (h NodeHandle) IsRoot() bool {
	return h.view.IsRoot(h.id)
}

// Parent returns a handle to the parent of the wrapped node, or a handle to
// [NilNodeID] if the node is a root.
func (h NodeHandle) Parent() NodeHandle {
	return h.view.Handle(h.view.Parent(h.id))
}

// NextSibling returns a handle to the next sibling of the wrapped node, or a
// handle to [NilNodeID] if the node has no further siblings.
func (h NodeHandle) NextSibling() NodeHandle {
	return h.view.Handle(h.view.NextSibling(h.id))
}

// FirstChild returns a handle to the first child of the wrapped node, or a
// handle to [NilNodeID] if the node has no children.
func (h NodeHandle) FirstChild() NodeHandle {
	return h.view.Handle(h.view.FirstChild(h.id))
}

// Detach detaches the wrapped node from its parent, turning it into a root node.
//
// It is equivalent to [NodeView.Detach].
func (h NodeHandle) Detach(preserveWorldTransform bool) {
	h.view.Detach(h.id, preserveWorldTransform)
}

// AttachChild attaches the specified child node as a child of the wrapped node.
//
// It is equivalent to [NodeView.AttachChild] with the wrapped node as the
// parent.
func (h NodeHandle) AttachChild(child NodeHandle, preserveWorldTransform bool) {
	h.view.AttachChild(h.id, child.ID(), preserveWorldTransform)
}

// Walk traverses the subtree rooted at the wrapped node, invoking the specified
// callback with a handle to each node.
//
// The traversal includes the wrapped node itself as well as all of its
// descendants, visiting each node before its children (depth-first, pre-order).
// The callback should return true to continue the traversal or false to stop it
// early. It is equivalent to iterating [NodeView.WalkSubtree] with the wrapped
// node as the root.
func (h NodeHandle) Walk(yield func(NodeHandle) bool) {
	for nodeID := range h.view.WalkSubtreeIter(h.id) {
		if !yield(h.view.Handle(nodeID)) {
			break
		}
	}
}

// WalkIter returns an iterator that traverses the subtree rooted at the wrapped
// node, visiting each node before its children.
//
// See [NodeHandle.Walk] for the traversal order.
func (h NodeHandle) WalkIter() iter.Seq[NodeHandle] {
	return func(yield func(NodeHandle) bool) {
		h.Walk(yield)
	}
}

// SubtreeContains returns whether the specified node lies within the subtree
// rooted at the wrapped node.
//
// The subtree includes the wrapped node itself, so a node contains itself. It
// is equivalent to [NodeView.SubtreeContains] with the wrapped node as the
// root.
func (h NodeHandle) SubtreeContains(find NodeHandle) bool {
	return h.view.SubtreeContains(h.id, find.ID())
}

// FindNode returns a handle to a node with the specified name within the subtree
// rooted at the wrapped node, or a handle to [NilNodeID] if no such node exists.
//
// The search includes the wrapped node itself as well as all of its
// descendants. It is equivalent to [NodeView.FindNodeInSubtree] with the wrapped
// node as the root.
func (h NodeHandle) FindNode(name string) NodeHandle {
	return h.view.Handle(h.view.FindNodeInSubtree(h.id, name))
}

// IsIndependent returns whether the wrapped node's transformation is
// independent of its parent.
//
// It is equivalent to [NodeView.IsIndependent].
func (h NodeHandle) IsIndependent() bool {
	return h.view.IsIndependent(h.id)
}

// SetIndependent sets whether the wrapped node's transformation is independent
// of its parent.
//
// It is equivalent to [NodeView.SetIndependent].
func (h NodeHandle) SetIndependent(independent, preserveWorldTransform bool) {
	h.view.SetIndependent(h.id, independent, preserveWorldTransform)
}

// IsHidden returns whether the wrapped node is explicitly hidden.
//
// It is equivalent to [NodeView.IsHidden].
func (h NodeHandle) IsHidden() bool {
	return h.view.IsHidden(h.id)
}

// IsVisible returns whether the wrapped node is not explicitly hidden.
//
// It is equivalent to [NodeView.IsVisible].
func (h NodeHandle) IsVisible() bool {
	return h.view.IsVisible(h.id)
}

// IsAbsoluteHidden returns whether the wrapped node is effectively hidden,
// either because it is hidden itself or because one of its ancestors is hidden.
//
// It is equivalent to [NodeView.IsAbsoluteHidden].
func (h NodeHandle) IsAbsoluteHidden() bool {
	return h.view.IsAbsoluteHidden(h.id)
}

// IsAbsoluteVisible returns whether the wrapped node is effectively visible,
// meaning that neither it nor any of its ancestors is hidden.
//
// It is equivalent to [NodeView.IsAbsoluteVisible].
func (h NodeHandle) IsAbsoluteVisible() bool {
	return h.view.IsAbsoluteVisible(h.id)
}

// SetHidden sets whether the wrapped node is hidden.
//
// It is equivalent to [NodeView.SetHidden].
func (h NodeHandle) SetHidden(hidden bool) {
	h.view.SetHidden(h.id, hidden)
}

// SetVisible sets whether the wrapped node is visible.
//
// It is equivalent to [NodeView.SetVisible].
func (h NodeHandle) SetVisible(visible bool) {
	h.view.SetVisible(h.id, visible)
}

// Name returns the name of the wrapped node.
//
// It is equivalent to [NodeView.Name].
func (h NodeHandle) Name() string {
	return h.view.Name(h.id)
}

// SetName sets the name of the wrapped node.
//
// It is equivalent to [NodeView.SetName].
func (h NodeHandle) SetName(name string) {
	h.view.SetName(h.id, name)
}

// Snap records the current absolute transformations of the wrapped node and all
// of its descendants as their previous transformations, so that teleporting the
// node does not cause rendering to interpolate across the jump.
//
// It is equivalent to [NodeView.Snap].
func (h NodeHandle) Snap() {
	h.view.Snap(h.id)
}

// Position returns the local position of the wrapped node, relative to its
// parent.
//
// It is equivalent to [NodeView.Position].
func (h NodeHandle) Position() dprec.Vec3 {
	return h.view.Position(h.id)
}

// SetPosition sets the local position of the wrapped node, relative to its
// parent.
//
// It is equivalent to [NodeView.SetPosition].
func (h NodeHandle) SetPosition(position dprec.Vec3) {
	h.view.SetPosition(h.id, position)
}

// Rotation returns the local rotation of the wrapped node, relative to its
// parent.
//
// It is equivalent to [NodeView.Rotation].
func (h NodeHandle) Rotation() dprec.Quat {
	return h.view.Rotation(h.id)
}

// SetRotation sets the local rotation of the wrapped node, relative to its
// parent.
//
// It is equivalent to [NodeView.SetRotation].
func (h NodeHandle) SetRotation(rotation dprec.Quat) {
	h.view.SetRotation(h.id, rotation)
}

// Scale returns the local scale of the wrapped node, relative to its parent.
//
// It is equivalent to [NodeView.Scale].
func (h NodeHandle) Scale() dprec.Vec3 {
	return h.view.Scale(h.id)
}

// SetScale sets the local scale of the wrapped node, relative to its parent.
//
// It is equivalent to [NodeView.SetScale].
func (h NodeHandle) SetScale(scale dprec.Vec3) {
	h.view.SetScale(h.id, scale)
}

// TRS returns the local translation, rotation, and scale of the wrapped node,
// relative to its parent.
//
// It is equivalent to [NodeView.TRS].
func (h NodeHandle) TRS() (dprec.Vec3, dprec.Quat, dprec.Vec3) {
	return h.view.TRS(h.id)
}

// SetTRS sets the local translation, rotation, and scale of the wrapped node,
// relative to its parent, in a single operation.
//
// It is equivalent to [NodeView.SetTRS].
func (h NodeHandle) SetTRS(position dprec.Vec3, rotation dprec.Quat, scale dprec.Vec3) {
	h.view.SetTRS(h.id, position, rotation, scale)
}

// Matrix returns the local transformation matrix of the wrapped node, relative
// to its parent.
//
// It is equivalent to [NodeView.Matrix].
func (h NodeHandle) Matrix() dprec.Mat4 {
	return h.view.Matrix(h.id)
}

// SetMatrix sets the local transformation of the wrapped node from the
// specified matrix, relative to its parent.
//
// It is equivalent to [NodeView.SetMatrix].
func (h NodeHandle) SetMatrix(matrix dprec.Mat4) {
	h.view.SetMatrix(h.id, matrix)
}

// ReferenceMatrix returns the absolute transformation matrix that serves as the
// reference frame for the wrapped node's local transformation.
//
// It is equivalent to [NodeView.ReferenceMatrix].
func (h NodeHandle) ReferenceMatrix() dprec.Mat4 {
	return h.view.ReferenceMatrix(h.id)
}

// AbsoluteMatrix returns the absolute (world) transformation matrix of the
// wrapped node.
//
// It is equivalent to [NodeView.AbsoluteMatrix].
func (h NodeHandle) AbsoluteMatrix() dprec.Mat4 {
	return h.view.AbsoluteMatrix(h.id)
}

// SetAbsoluteMatrix sets the local transformation of the wrapped node so that
// its absolute (world) matrix equals the specified matrix.
//
// It is equivalent to [NodeView.SetAbsoluteMatrix].
func (h NodeHandle) SetAbsoluteMatrix(matrix dprec.Mat4) {
	h.view.SetAbsoluteMatrix(h.id, matrix)
}

// InterpolatedAbsoluteMatrix returns the wrapped node's absolute (world) matrix
// interpolated between its previous and current values by the specified
// fraction.
//
// It is equivalent to [NodeView.InterpolatedAbsoluteMatrix].
func (h NodeHandle) InterpolatedAbsoluteMatrix(fraction float64) dprec.Mat4 {
	return h.view.InterpolatedAbsoluteMatrix(h.id, fraction)
}

const (
	nodeFlagDirty int32 = 1 << iota
	nodeFlagIndependent
	nodeFlagParentHidden
	nodeFlagHidden
)

type nodeState struct {
	parentIndex      int32
	nextSiblingIndex int32
	firstChildIndex  int32

	revision int32
	flags    int32

	name string

	position     dprec.Vec3
	rotation     dprec.Quat
	scale        dprec.Vec3
	absMatrix    dprec.Mat4
	oldAbsMatrix dprec.Mat4
}

func (s nodeState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}

func (n *nodeState) hasFlag(flag int32) bool {
	return n.flags&flag != 0
}

func (n *nodeState) setFlag(flag int32) {
	n.flags |= flag
}

func (n *nodeState) unsetFlag(flag int32) {
	n.flags &^= flag
}

func (n *nodeState) isTransformIndependent() bool {
	return n.hasFlag(nodeFlagIndependent) || (n.parentIndex == nilIndex)
}

func (n *nodeState) isAbsoluteHidden() bool {
	return n.hasFlag(nodeFlagHidden) || n.hasFlag(nodeFlagParentHidden)
}

func (n *nodeState) calculateMatrix() dprec.Mat4 {
	return dprec.TRSMat4(n.position, n.rotation, n.scale)
}
