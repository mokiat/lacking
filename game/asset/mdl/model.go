package mdl

import "slices"

type Model struct {
	name string

	nodes      []Node
	animations []*Animation
}

func (s *Model) Name() string {
	return s.name
}

func (s *Model) SetName(name string) {
	s.name = name
}

func (s *Model) Nodes() []Node {
	return s.nodes
}

func (s *Model) AddNode(node Node) {
	s.nodes = append(s.nodes, node)
}

func (s *Model) RemoveNode(node Node) {
	s.nodes = slices.DeleteFunc(s.nodes, func(candidate Node) bool {
		return candidate == node
	})
}

func (s *Model) FlattenNodes() []Node {
	var nodes []Node
	var visit func(Node)
	visit = func(n Node) {
		nodes = append(nodes, n)
		for _, child := range n.Nodes() {
			visit(child)
		}
	}
	for _, node := range s.nodes {
		visit(node)
	}
	return nodes
}

func (s *Model) Animations() []*Animation {
	return s.animations
}

func (s *Model) AddAnimation(animation *Animation) {
	s.animations = append(s.animations, animation)
}
