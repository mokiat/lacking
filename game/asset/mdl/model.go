package mdl

import (
	"iter"
	"slices"

	"github.com/mokiat/gog/ds"
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

func (s *Model) AllPhysicsBodyMaterials() []*BodyMaterial {
	var result []*BodyMaterial
	seen := ds.NewSet[*BodyMaterial](0)
	for _, definition := range s.AllPhysicsBodyDefinitions() {
		material := definition.Material()
		if !seen.Contains(material) {
			result = append(result, material)
		}
		seen.Add(material)
	}
	return result
}

func (s *Model) AllPhysicsBodyDefinitions() []*BodyDefinition {
	var result []*BodyDefinition
	seen := ds.NewSet[*BodyDefinition](0)
	for _, placement := range s.AllPhysicsBodyPlacements() {
		body := placement.Value
		definition := body.Definition()
		if !seen.Contains(definition) {
			result = append(result, definition)
		}
		seen.Add(definition)
	}
	return result
}

func (s *Model) AllPhysicsBodyPlacements() []Placed[*Body] {
	var result []Placed[*Body]
	for _, node := range s.NodesIter() {
		switch source := node.Source().(type) {
		case *Body:
			result = append(result, Placed[*Body]{
				Node:  node,
				Value: source,
			})
		}
	}
	return result
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
