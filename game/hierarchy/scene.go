package hierarchy

import (
	"iter"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
)

// NewScene creates a new scene with the specified initial capacity for nodes.
func NewScene(initialCapacity int) *Scene {
	return &Scene{
		deleteSubscriptions: NewDeleteSubscriptionSet(),

		nodes:           make([]node, 0, initialCapacity),
		freeNodeIndices: ds.NewStack[int32](initialCapacity),
		freeRevision:    1,
		freeTimestamp:   1,
	}
}

type Scene struct {
	deleteSubscriptions *DeleteSubscriptionSet

	nodes           []node
	freeNodeIndices *ds.Stack[int32]
	freeRevision    uint32
	freeTimestamp   int32
}

// SubscribeNodeDelete subscribes to node deletion events.
func (s *Scene) SubscribeNodeDelete(callback DeleteCallback) *DeleteSubscription {
	return s.deleteSubscriptions.Subscribe(callback)
}

// Wrap creates a Node wrapper for the specified NodeID.
func (s *Scene) Wrap(id NodeID) Node {
	return Node{
		scene: s,
		id:    id,
	}
}

// CreateNode creates a new node in the scene and returns its ID.
func (s *Scene) CreateNode() NodeID {
	var index int32
	if s.freeNodeIndices.IsEmpty() {
		index = int32(len(s.nodes))
		if len(s.nodes) < cap(s.nodes) {
			s.nodes = s.nodes[:index+1]
		} else {
			s.nodes = append(s.nodes, node{})
		}
	} else {
		index = s.freeNodeIndices.Pop()
	}
	node := &s.nodes[index]
	node.initialize(index, s.nextRevision())
	return NodeID{
		index:    index,
		revision: node.revision,
	}
}

// DeleteNode deletes the node with the specified ID from the scene.
func (s *Scene) DeleteNode(id NodeID) {
	if !s.IsValidNode(id) {
		return
	}
	s.freeNode(id.index)
}

// IsValidNode returns whether the specified ID represents a valid node in the
// scene.
func (s *Scene) IsValidNode(id NodeID) bool {
	if id.revision == 0 {
		return false
	}
	return s.nodes[id.index].revision == id.revision
}

// NodeName returns the name of the node with the specified ID.
func (s *Scene) NodeName(id NodeID) string {
	node := s.fetchNode(id)
	return node.getName(s)
}

// SetNodeName sets the name of the node with the specified ID.
func (s *Scene) SetNodeName(id NodeID, name string) {
	node := s.fetchNode(id)
	node.setName(s, name)
}

// IsNodeVisible returns whether the node with the specified ID is marked as
// visible.
func (s *Scene) IsNodeVisible(id NodeID) bool {
	node := s.fetchNode(id)
	return node.getIsVisible(s)
}

// SetNodeVisible sets whether the node with the specified ID is marked as
// visible.
func (s *Scene) SetNodeVisible(id NodeID, visible bool) {
	node := s.fetchNode(id)
	node.setIsVisible(s, visible)
}

// IsNodeIndependent returns whether the node with the specified ID is marked as
// independent (i.e. not affected by parent transformations).
func (s *Scene) IsNodeIndependent(id NodeID) bool {
	node := s.fetchNode(id)
	return node.getIsIndependent(s)
}

// SetNodeIndependent sets whether the node with the specified ID is marked as
// independent (i.e. not affected by parent transformations).
func (s *Scene) SetNodeIndependent(id NodeID, independent bool) {
	node := s.fetchNode(id)
	node.setIsIndependent(s, independent)
}

// NodeTransformFunc returns the custom transformation function of the node with
// the specified ID.
func (s *Scene) NodeTransformFunc(id NodeID) TransformFunc {
	node := s.fetchNode(id)
	return node.getTransform(s)
}

// SetNodeTransformFunc sets the custom transformation function of the node with
// the specified ID.
func (s *Scene) SetNodeTransformFunc(id NodeID, transformFunc TransformFunc) {
	node := s.fetchNode(id)
	node.setTransform(s, transformFunc)
}

// NodeHasParent returns whether the node with the specified ID has a parent.
func (s *Scene) NodeHasParent(id NodeID) bool {
	node := s.fetchNode(id)
	return node.parentIndex != -1
}

// NodeParent returns the ID of the parent of the node with the specified ID.
func (s *Scene) NodeParent(id NodeID) NodeID {
	node := s.fetchNode(id)
	if node.parentIndex == -1 {
		return NilNodeID
	}
	parent := &s.nodes[node.parentIndex]
	return parent.getID()
}

// NodeFirstChild returns the ID of the first child of the node with the
// specified ID.
func (s *Scene) NodeFirstChild(id NodeID) NodeID {
	node := s.fetchNode(id)
	if node.firstChildIndex == -1 {
		return NilNodeID
	}
	child := &s.nodes[node.firstChildIndex]
	return child.getID()
}

// NodeLastChild returns the ID of the last child of the node with the
// specified ID.
func (s *Scene) NodeLastChild(id NodeID) NodeID {
	node := s.fetchNode(id)
	if node.lastChildIndex == -1 {
		return NilNodeID
	}
	child := &s.nodes[node.lastChildIndex]
	return child.getID()
}

// NodeLeftSibling returns the ID of the left sibling of the node with the
// specified ID.
func (s *Scene) NodeLeftSibling(id NodeID) NodeID {
	node := s.fetchNode(id)
	if node.leftSiblingIndex == -1 {
		return NilNodeID
	}
	sibling := &s.nodes[node.leftSiblingIndex]
	return sibling.getID()
}

// NodeRightSibling returns the ID of the right sibling of the node with the
// specified ID.
func (s *Scene) NodeRightSibling(id NodeID) NodeID {
	node := s.fetchNode(id)
	if node.rightSiblingIndex == -1 {
		return NilNodeID
	}
	sibling := &s.nodes[node.rightSiblingIndex]
	return sibling.getID()
}

// ResetDelta resets the delta state of all nodes in the scene.
func (s *Scene) ResetDelta() {
	for i := range s.nodes {
		node := &s.nodes[i]
		node.resetDelta(s, false)
	}
}

// ResetNodeDelta resets the delta state of the node with the specified ID.
//
// If recursive is true, the delta state of all descendant nodes is also reset.
func (s *Scene) ResetNodeDelta(id NodeID, recursive bool) {
	node := s.fetchNode(id)
	node.resetDelta(s, recursive)
}

// FindNode returns the first found node with the specified name or
// NilNodeID if no such node exists.
//
// The search order is arbitrary and should not be relied upon, hence use
// this method for nodes that have unique names only.
func (s *Scene) FindNode(name string) NodeID {
	for i := range s.nodes {
		node := &s.nodes[i]
		if node.revision == 0 {
			continue
		}
		if node.name == name {
			return NodeID{
				index:    int32(i),
				revision: node.revision,
			}
		}
	}
	return NilNodeID
}

// FindSubtreeNode returns the first found node with the specified name in the
// subtree rooted at the node with the specified ID or NilNodeID if no such node
// exists.
func (s *Scene) FindSubtreeNode(rootID NodeID, name string) NodeID {
	if rootID == NilNodeID {
		return s.FindNode(name)
	}
	if !s.IsValidNode(rootID) {
		return NilNodeID
	}
	index := s.findSubtreeNode(rootID.index, name)
	if index == -1 {
		return NilNodeID
	}
	node := &s.nodes[index]
	return NodeID{
		index:    index,
		revision: node.revision,
	}
}

// IsNodeChain returns whether the node with the specified leaf ID is a
// descendant of the node with the specified root ID.
func (s *Scene) IsNodeChain(rootID, leafID NodeID) bool {
	if !s.IsValidNode(rootID) || !s.IsValidNode(leafID) {
		return false
	}
	if rootID == leafID {
		return true
	}
	current := s.fetchNode(leafID)
	for current.parentIndex != -1 {
		if current.parentIndex == rootID.index {
			return true
		}
		current = &s.nodes[current.parentIndex]
	}
	return false
}

// PrependNodeSibling attaches the node with the specified siblingID to the left
// of the node with the specified ID.
func (s *Scene) PrependNodeSibling(currentID, siblingID NodeID, preserveWorldTransform bool) {
	currentNode := s.fetchNode(currentID)
	siblingNode := s.fetchNode(siblingID)
	if preserveWorldTransform {
		siblingNode.ensureNotStale(s)
	}
	siblingNode.detach(s)
	currentNode.prependSibling(s, siblingNode)
	if preserveWorldTransform {
		siblingNode.reconstructTransforms(s)
	}
}

// AppendNodeSibling attaches the node with the specified siblingID to the right
// of the node with the specified ID.
func (s *Scene) AppendNodeSibling(currentID, siblingID NodeID, preserveWorldTransform bool) {
	currentNode := s.fetchNode(currentID)
	siblingNode := s.fetchNode(siblingID)
	if preserveWorldTransform {
		siblingNode.ensureNotStale(s)
	}
	siblingNode.detach(s)
	currentNode.appendSibling(s, siblingNode)
	if preserveWorldTransform {
		siblingNode.reconstructTransforms(s)
	}
}

// PrependNodeChild attaches the node with the specified childID as the first
// child of the node with the specified parentID.
func (s *Scene) PrependNodeChild(parentID, childID NodeID, preserveWorldTransform bool) {
	parentNode := s.fetchNode(parentID)
	childNode := s.fetchNode(childID)
	if preserveWorldTransform {
		childNode.ensureNotStale(s)
	}
	childNode.detach(s)
	parentNode.prependChild(s, childNode)
	if preserveWorldTransform {
		childNode.reconstructTransforms(s)
	}
}

// AppendNodeChild attaches the node with the specified childID as the last
// child of the node with the specified parentID.
func (s *Scene) AppendNodeChild(parentID, childID NodeID, preserveWorldTransform bool) {
	parentNode := s.fetchNode(parentID)
	childNode := s.fetchNode(childID)
	if preserveWorldTransform {
		childNode.ensureNotStale(s)
	}
	childNode.detach(s)
	parentNode.appendChild(s, childNode)
	if preserveWorldTransform {
		childNode.reconstructTransforms(s)
	}
}

// DetachNode detaches the node with the specified ID from its parent.
func (s *Scene) DetachNode(id NodeID, preserveWorldTransform bool) {
	node := s.fetchNode(id)
	if preserveWorldTransform {
		node.ensureNotStale(s)
	}
	node.detach(s)
	if preserveWorldTransform {
		node.reconstructTransforms(s)
	}
}

// NodePosition returns the local position of the node with the specified ID.
func (s *Scene) NodePosition(id NodeID) dprec.Vec3 {
	node := s.fetchNode(id)
	return node.getPosition(s)
}

// SetNodePosition sets the local position of the node with the specified ID.
func (s *Scene) SetNodePosition(id NodeID, position dprec.Vec3) {
	node := s.fetchNode(id)
	node.setPosition(s, position)
}

// InitializeNodePosition initializes the local position of the node with the
// specified ID, avoiding interpolation with previous value.
//
// If multiple aspects are initialized, it is best to use the Set methods
// followed by a ResetNodeDelta call.
func (s *Scene) InitializeNodePosition(id NodeID, position dprec.Vec3) {
	s.SetNodePosition(id, position)
	s.ResetNodeDelta(id, true)
}

// NodeRotation returns the local rotation of the node with the specified ID.
func (s *Scene) NodeRotation(id NodeID) dprec.Quat {
	node := s.fetchNode(id)
	return node.getRotation(s)
}

// SetNodeRotation sets the local rotation of the node with the specified ID.
func (s *Scene) SetNodeRotation(id NodeID, rotation dprec.Quat) {
	node := s.fetchNode(id)
	node.setRotation(s, rotation)
}

// InitializeNodeRotation initializes the local rotation of the node with the
// specified ID, avoiding interpolation with previous value.
//
// If multiple aspects are initialized, it is best to use the Set methods
// followed by a ResetNodeDelta call.
func (s *Scene) InitializeNodeRotation(id NodeID, rotation dprec.Quat) {
	s.SetNodeRotation(id, rotation)
	s.ResetNodeDelta(id, false)
}

// NodeScale returns the local scale of the node with the specified ID.
func (s *Scene) NodeScale(id NodeID) dprec.Vec3 {
	node := s.fetchNode(id)
	return node.getScale(s)
}

// SetNodeScale sets the local scale of the node with the specified ID.
func (s *Scene) SetNodeScale(id NodeID, scale dprec.Vec3) {
	node := s.fetchNode(id)
	node.setScale(s, scale)
}

// InitializeNodeScale initializes the local scale of the node with the
// specified ID, avoiding interpolation with previous value.
//
// If multiple aspects are initialized, it is best to use the Set methods
// followed by a ResetNodeDelta call.
func (s *Scene) InitializeNodeScale(id NodeID, scale dprec.Vec3) {
	s.SetNodeScale(id, scale)
	s.ResetNodeDelta(id, false)
}

// NodeMatrix returns the local transformation matrix of the node with the
// specified ID.
func (s *Scene) NodeMatrix(id NodeID) dprec.Mat4 {
	node := s.fetchNode(id)
	return node.getMatrix(s)
}

// SetNodeMatrix sets the local transformation matrix of the node with the
// specified ID.
func (s *Scene) SetNodeMatrix(id NodeID, matrix dprec.Mat4) {
	node := s.fetchNode(id)
	node.setMatrix(s, matrix)
}

// InitializeNodeMatrix initializes the local transformation matrix of the
// node with the specified ID, avoiding interpolation with previous value.
func (s *Scene) InitializeNodeMatrix(id NodeID, matrix dprec.Mat4) {
	s.SetNodeMatrix(id, matrix)
	s.ResetNodeDelta(id, false)
}

// NodeBaseMatrix returns the base transformation matrix of the node with the
// specified ID - this is the absolute matrix of the parent, if there is one,
// or the identity matrix otherwise.
func (s *Scene) NodeBaseMatrix(id NodeID) dprec.Mat4 {
	node := s.fetchNode(id)
	if node.isIndependent || node.parentIndex == -1 {
		return dprec.IdentityMat4()
	}
	parent := &s.nodes[node.parentIndex]
	return parent.getAbsoluteMatrix(s)
}

// NodeAbsoluteMatrix returns the absolute transformation matrix of the node
// with the specified ID.
func (s *Scene) NodeAbsoluteMatrix(id NodeID) dprec.Mat4 {
	node := s.fetchNode(id)
	return node.getAbsoluteMatrix(s)
}

// SetNodeAbsoluteMatrix sets the absolute transformation matrix of the node
// with the specified ID.
func (s *Scene) SetNodeAbsoluteMatrix(id NodeID, matrix dprec.Mat4) {
	node := s.fetchNode(id)
	if node.isIndependent || node.parentIndex == -1 {
		node.setMatrix(s, matrix)
	} else {
		parent := &s.nodes[node.parentIndex]
		parentMatrix := parent.getAbsoluteMatrix(s)
		relativeMatrix := dprec.Mat4Prod(
			dprec.InverseMat4(parentMatrix),
			matrix,
		)
		node.setMatrix(s, relativeMatrix)
	}
}

// NodeInterpolatedAbsoluteMatrix returns the absolute transformation matrix of
// the node with the specified ID, interpolated between the previous and
// current states by the specified fraction in [0, 1].
func (s *Scene) NodeInterpolatedAbsoluteMatrix(id NodeID, fraction float64) dprec.Mat4 {
	node := s.fetchNode(id)
	prevTranslation, prevRotation, prevScale := node.getPreviousAbsoluteTRS(s)
	currTranslation, currRotation, currScale := node.getCurrentAbsoluteTRS(s)
	return dprec.TRSMat4(
		dprec.Vec3Lerp(prevTranslation, currTranslation, fraction),
		dprec.QuatSlerp(prevRotation, currRotation, fraction),
		dprec.Vec3Lerp(prevScale, currScale, fraction),
	)
}

// Visit traverses the all the nodes in a depth-first manner, invoking the
// specified callback for each node in the scene.
func (s *Scene) Visit(callback func(NodeID) bool) {
	for _, node := range s.nodes {
		if node.isDeleted {
			continue
		}
		if node.parentIndex == -1 {
			if !s.yieldSubtree(node.index, callback) {
				return
			}
		}
	}
}

// VisitSubtree traverses the subtree rooted at the node with the specified ID,
// invoking the specified callback for each node in the subtree.
func (s *Scene) VisitSubtree(rootID NodeID, callback func(NodeID) bool) {
	_ = s.yieldSubtree(rootID.index, callback)
}

// SubtreeIter returns an iterator that traverses the subtree rooted at the node
// with the specified ID.
func (s *Scene) SubtreeIter(rootID NodeID) iter.Seq[NodeID] {
	return func(yield func(NodeID) bool) {
		_ = s.yieldSubtree(rootID.index, yield)
	}
}

func (s *Scene) yieldSubtree(index int32, callback func(NodeID) bool) bool {
	node := &s.nodes[index]
	if !callback(node.getID()) {
		return false
	}
	for childIndex := node.firstChildIndex; childIndex != -1; {
		child := &s.nodes[childIndex]
		if !s.yieldSubtree(childIndex, callback) {
			return false
		}
		childIndex = child.rightSiblingIndex
	}
	return true
}

func (s *Scene) nextRevision() uint32 {
	revision := s.freeRevision
	s.freeRevision++
	return revision
}

func (s *Scene) nextTimestamp() int32 {
	timestamp := s.freeTimestamp
	s.freeTimestamp++
	return timestamp
}

func (s *Scene) fetchNode(id NodeID) *node {
	if id.revision == 0 {
		panic("invalid or deleted node")
	}
	node := &s.nodes[id.index]
	if node.revision != id.revision {
		panic("invalid or deleted node")
	}
	return node
}

func (s *Scene) findSubtreeNode(index int32, name string) int32 {
	node := &s.nodes[index]
	if node.name == name {
		return index
	}
	for childIndex := node.firstChildIndex; childIndex != -1; {
		foundIndex := s.findSubtreeNode(childIndex, name)
		if foundIndex != -1 {
			return foundIndex
		}
		child := &s.nodes[childIndex]
		childIndex = child.rightSiblingIndex
	}
	return -1
}

func (s *Scene) freeNode(index int32) {
	node := &s.nodes[index]

	childIndex := node.firstChildIndex
	for childIndex != -1 {
		child := &s.nodes[childIndex]
		nextChildIndex := child.rightSiblingIndex
		s.freeNode(childIndex)
		childIndex = nextChildIndex
	}

	s.deleteSubscriptions.Each(func(callback DeleteCallback) {
		callback(s, node.getID())
	})

	node.detach(s)
	node.revision++
	node.isDeleted = true
	s.freeNodeIndices.Push(index)
}
