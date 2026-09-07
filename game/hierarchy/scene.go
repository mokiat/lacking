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

// ApplySourcesToNodes runs the source transfer across every binding in the
// scene, updating each bound node from its bound value.
//
// For every binding that has a [SourceBindingSolver], its
// [SourceBindingSolver.OnSourceToNode] is invoked for each node bound in it.
// Bindings are processed in ascending [Binding.Priority] order.
func (s *Scene) ApplySourcesToNodes() {
	for _, binding := range s.bindings {
		binding.ApplySourcesToNodes()
	}
}

// ApplySourceToNode runs the source transfer for a single node across every
// binding in the scene, updating the node from its bound values.
//
// If recursive is true, the transfer is applied to the node and all of its
// descendants; otherwise only to the node itself. Bindings are processed in
// ascending [Binding.Priority] order.
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

// ApplyTargetsFromNodes runs the target transfer across every binding in the
// scene, updating each bound value from its node.
//
// For every binding that has a [TargetBindingSolver], its
// [TargetBindingSolver.OnTargetFromNode] is invoked for each node bound in it.
// Bindings are processed in ascending [Binding.Priority] order.
func (s *Scene) ApplyTargetsFromNodes() {
	for _, binding := range s.bindings {
		binding.ApplyTargetsFromNodes()
	}
}

// ApplyTargetFromNode runs the target transfer for a single node across every
// binding in the scene, updating the node's bound values from the node.
//
// If recursive is true, the transfer is applied to the node and all of its
// descendants; otherwise only to the node itself. Bindings are processed in
// ascending [Binding.Priority] order.
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

// ApplyInterpolationsFromNodes runs the interpolation transfer across every
// binding in the scene, updating each bound value from its node's pose
// interpolated by the specified fraction.
//
// For every binding that has an [InterpolationBindingSolver], its
// [InterpolationBindingSolver.OnInterpolationFromNode] is invoked for each node
// bound in it. Bindings are processed in ascending [Binding.Priority] order.
func (s *Scene) ApplyInterpolationsFromNodes(fraction float64) {
	for _, binding := range s.bindings {
		binding.ApplyInterpolationsFromNodes(fraction)
	}
}

// ApplyInterpolationFromNode runs the interpolation transfer for a single node
// across every binding in the scene, updating the node's bound values from the
// node's pose interpolated by the specified fraction.
//
// If recursive is true, the transfer is applied to the node and all of its
// descendants; otherwise only to the node itself. Bindings are processed in
// ascending [Binding.Priority] order.
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

func (s *Scene) addBinding(binding internalBinding) {
	s.bindings = append(s.bindings, binding)
	s.sortBindings()
}

func (s *Scene) removeBinding(binding internalBinding) {
	s.bindings = slices.DeleteFunc(s.bindings, func(b internalBinding) bool {
		return b == binding
	})
}

func (s *Scene) sortBindings() {
	slices.SortStableFunc(s.bindings, func(a, b internalBinding) int {
		return cmp.Compare(a.Priority(), b.Priority())
	})
}

const nilIndex int32 = -1
