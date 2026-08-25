package hierarchy

import (
	"cmp"
	"slices"

	"github.com/mokiat/gog/ds"
)

// Scene represents a hierarchy of nodes, each of which can have a
// transformation and can be organized in a parent-child relationship.
type Scene struct {
	freeNodeIndices *ds.Stack[int32]

	nodes    []nodeState
	bindings []internalBinding
}

// NewScene creates a new empty [Scene].
func NewScene() *Scene {
	return &Scene{
		freeNodeIndices: ds.EmptyStack[int32](),

		nodes:    make([]nodeState, 0),
		bindings: make([]internalBinding, 0),
	}
}

// Nodes returns a [NodeView] that provides access to the nodes in the scene,
// allowing them to be created, queried, and manipulated.
func (s *Scene) Nodes() NodeView {
	return NodeView{
		scene: s,
	}
}

// AdvanceStep records the current absolute transformations of all nodes in the
// scene as their previous transformations.
//
// The recorded transformations become the starting point that
// [NodeView.InterpolatedAbsoluteMatrix] interpolates from towards each node's
// current transformation. Call this once per fixed-rate update, before applying
// the new transformations for that step, so that rendering can interpolate
// smoothly between the previous and the current pose.
//
// To teleport an individual node without rendering interpolating across the
// jump, use [NodeView.Snap] instead.
func (s *Scene) AdvanceStep() {
	nodeView := s.Nodes()
	for i := range s.nodes {
		node := &s.nodes[i]
		if node.isValid() {
			nodeView.snapNode(node, false)
		}
	}
}

func (s *Scene) ApplySourcesToNodes() {
	for _, binding := range s.bindings {
		binding.ApplySourcesToNodes()
	}
}

func (s *Scene) ApplySourceToNode(id NodeID, recursive bool) {
	if recursive {
		s.Nodes().WalkSubtree(id, func(nodeID NodeID) bool {
			s.ApplySourceToNode(nodeID, false)
			return true
		})
		return
	}

	for _, binding := range s.bindings {
		binding.ApplySourceToNode(id)
	}
}

func (s *Scene) ApplyTargetsFromNodes() {
	for _, binding := range s.bindings {
		binding.ApplyTargetsFromNodes()
	}
}

func (s *Scene) ApplyTargetFromNode(id NodeID, recursive bool) {
	if recursive {
		s.Nodes().WalkSubtree(id, func(nodeID NodeID) bool {
			s.ApplyTargetFromNode(nodeID, false)
			return true
		})
		return
	}

	for _, binding := range s.bindings {
		binding.ApplyTargetFromNode(id)
	}
}

func (s *Scene) ApplyInterpolationsFromNodes(fraction float64) {
	for _, binding := range s.bindings {
		binding.ApplyInterpolationsFromNodes(fraction)
	}
}

func (s *Scene) ApplyInterpolationFromNode(id NodeID, fraction float64, recursive bool) {
	if recursive {
		s.Nodes().WalkSubtree(id, func(nodeID NodeID) bool {
			s.ApplyInterpolationFromNode(nodeID, fraction, false)
			return true
		})
		return
	}

	for _, binding := range s.bindings {
		binding.ApplyInterpolationFromNode(id, fraction)
	}
}

func (s *Scene) allocateNode() (int32, *nodeState) {
	var index int32
	if s.freeNodeIndices.IsEmpty() {
		index = int32(len(s.nodes))
		s.nodes = append(s.nodes, nodeState{})
	} else {
		index = s.freeNodeIndices.Pop()
	}
	return index, &s.nodes[index]
}

func (s *Scene) releaseNode(index int32) {
	s.freeNodeIndices.Push(index)
}

func (s *Scene) addBindingSet(binding internalBinding) {
	s.bindings = append(s.bindings, binding)
	s.sortBindingSets()
}

func (s *Scene) removeBindingSet(binding internalBinding) {
	s.bindings = slices.DeleteFunc(s.bindings, func(b internalBinding) bool {
		return b == binding
	})
}

func (s *Scene) sortBindingSets() {
	slices.SortStableFunc(s.bindings, func(a, b internalBinding) int {
		return cmp.Compare(a.Priority(), b.Priority())
	})
}

const nilIndex int32 = -1
