package hierarchy2

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
)

func NewScene() *Scene {
	initialCapacity := 1024

	freeNodeIndices := ds.NewStack[int32](initialCapacity)
	for i := range int32(initialCapacity) {
		freeNodeIndices.Push(i)
	}
	return &Scene{
		nodes:           make([]node, 1024),
		freeNodeIndices: freeNodeIndices,
	}
}

type Scene struct {
	nodes           []node
	freeNodeIndices *ds.Stack[int32]
}

// CreateNode creates a new node in the scene and returns its ID.
func (s *Scene) CreateNode() NodeID {
	// TODO: Handle growing.
	index := s.freeNodeIndices.Pop()
	node := &s.nodes[index]
	node.revision++
	node.timestamp = initialTimestamp
	node.name = ""
	node.parentIndex = -1
	node.firstChildIndex = -1
	node.lastChildIndex = -1
	node.leftSiblingIndex = -1
	node.rightSiblingIndex = -1
	node.previousPosition = dprec.ZeroVec3()
	node.previousRotation = dprec.IdentityQuat()
	node.previousScale = dprec.NewVec3(1.0, 1.0, 1.0)
	node.previousAbsMatrix = dprec.IdentityMat4()
	node.currentPosition = dprec.ZeroVec3()
	node.currentRotation = dprec.IdentityQuat()
	node.currentScale = dprec.NewVec3(1.0, 1.0, 1.0)
	node.currentAbsMatrix = dprec.IdentityMat4()
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
	if id == NilNodeID {
		return false
	}
	return s.nodes[id.index].revision == id.revision
}

// NodeName returns the name of the node with the specified ID.
func (s *Scene) NodeName(id NodeID) string {
	node := s.fetchNode(id)
	return node.name
}

// SetNodeName sets the name of the node with the specified ID.
func (s *Scene) SetNodeName(id NodeID, name string) {
	node := s.fetchNode(id)
	node.name = name
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
	return NodeID{
		index:    node.parentIndex,
		revision: parent.revision,
	}
}

// NodeFirstChild returns the ID of the first child of the node with the
// specified ID.
func (s *Scene) NodeFirstChild(id NodeID) NodeID {
	node := s.fetchNode(id)
	if node.firstChildIndex == -1 {
		return NilNodeID
	}
	child := &s.nodes[node.firstChildIndex]
	return NodeID{
		index:    node.firstChildIndex,
		revision: child.revision,
	}
}

// NodeLastChild returns the ID of the last child of the node with the
// specified ID.
func (s *Scene) NodeLastChild(id NodeID) NodeID {
	node := s.fetchNode(id)
	if node.lastChildIndex == -1 {
		return NilNodeID
	}
	child := &s.nodes[node.lastChildIndex]
	return NodeID{
		index:    node.lastChildIndex,
		revision: child.revision,
	}
}

// NodeLeftSibling returns the ID of the left sibling of the node with the
// specified ID.
func (s *Scene) NodeLeftSibling(id NodeID) NodeID {
	node := s.fetchNode(id)
	if node.leftSiblingIndex == -1 {
		return NilNodeID
	}
	sibling := &s.nodes[node.leftSiblingIndex]
	return NodeID{
		index:    node.leftSiblingIndex,
		revision: sibling.revision,
	}
}

// func (s *Scene)

// NodeRightSibling returns the ID of the right sibling of the node with the
// specified ID.

func (s *Scene) fetchNode(id NodeID) *node {
	if id == NilNodeID {
		panic("invalid node id")
	}
	node := &s.nodes[id.index]
	if node.revision != id.revision {
		panic("node already deleted")
	}
	return node
}

func (s *Scene) freeNode(index int32) {
	s.freeNodeIndices.Push(index)
	node := &s.nodes[index]
	node.revision++

	childIndex := node.firstChildIndex
	for childIndex != -1 {
		child := &s.nodes[childIndex]
		nextChildIndex := child.rightSiblingIndex
		s.freeNode(childIndex)
		childIndex = nextChildIndex
	}
}
