package model

import "slices"

type Scene struct {
	name string

	nodes []Node
}

func (s *Scene) Name() string {
	return s.name
}

func (s *Scene) SetName(name string) {
	s.name = name
}

func (s *Scene) Nodes() []Node {
	return s.nodes
}

func (s *Scene) AddNode(node Node) {
	s.nodes = append(s.nodes, node)
}

func (s *Scene) RemoveNode(node Node) {
	s.nodes = slices.DeleteFunc(s.nodes, func(candidate Node) bool {
		return candidate == node
	})
}

func (s *Scene) FlattenNodes() []Node {
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
