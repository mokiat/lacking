package mdl

import (
	"iter"
	"slices"
)

type Model struct {
	name string

	nodes      []*Node
	animations []*Animation
}

func (s *Model) Name() string {
	return s.name
}

func (s *Model) SetName(name string) {
	s.name = name
}

func (s *Model) Nodes() []*Node {
	return s.nodes
}

func (s *Model) AddNode(node *Node) {
	s.nodes = append(s.nodes, node)
}

func (s *Model) RemoveNode(node *Node) {
	s.nodes = slices.DeleteFunc(s.nodes, func(candidate *Node) bool {
		return candidate == node
	})
}

func (s *Model) NodesIter() iter.Seq2[int, *Node] {
	return indexedIter(func(yield func(*Node) bool) {
		s.yieldNodes(s.nodes, yield)
	})
}

func (s *Model) yieldNodes(nodes []*Node, yield func(*Node) bool) bool {
	for _, node := range nodes {
		if !yield(node) {
			return false
		}
		if !s.yieldNodes(node.Nodes(), yield) {
			return false
		}
	}
	return true
}

func (s *Model) Animations() []*Animation {
	return s.animations
}

func (s *Model) AddAnimation(animation *Animation) {
	s.animations = append(s.animations, animation)
}

// TODO: Consider moving this to gog project.
func indexedIter[T any](src iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		index := 0
		for item := range src {
			if !yield(index, item) {
				return
			}
			index++
		}
	}
}
