package mdl

import (
	"slices"

	"github.com/mokiat/gomath/dprec"
)

func NewNode(name string) *Node {
	return &Node{
		name:        name,
		translation: dprec.ZeroVec3(),
		rotation:    dprec.IdentityQuat(),
		scale:       dprec.NewVec3(1.0, 1.0, 1.0),
	}
}

type Node struct {
	name string

	source any
	target any

	translation dprec.Vec3
	rotation    dprec.Quat
	scale       dprec.Vec3

	parent *Node
	nodes  []*Node
}

func (n *Node) Source() any {
	return n.source
}

func (n *Node) SetSource(source any) {
	n.source = source
}

func (n *Node) Target() any {
	return n.target
}

func (n *Node) SetTarget(target any) {
	n.target = target
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) SetName(name string) {
	n.name = name
}

func (n *Node) Translation() dprec.Vec3 {
	return n.translation
}

func (n *Node) SetTranslation(translation dprec.Vec3) {
	n.translation = translation
}

func (n *Node) Rotation() dprec.Quat {
	return n.rotation
}

func (n *Node) SetRotation(rotation dprec.Quat) {
	n.rotation = rotation
}

func (n *Node) Scale() dprec.Vec3 {
	return n.scale
}

func (n *Node) SetScale(scale dprec.Vec3) {
	n.scale = scale
}

func (n *Node) Parent() *Node {
	return n.parent
}

func (n *Node) SetParent(parent *Node) {
	n.parent = parent
}

func (n *Node) Nodes() []*Node {
	return n.nodes
}

func (n *Node) AddNode(node *Node) {
	node.SetParent(n)
	n.nodes = append(n.nodes, node)
}

func (n *Node) RemoveNode(node *Node) {
	if node.Parent() == n {
		node.SetParent(nil)
	}
	n.nodes = slices.DeleteFunc(n.nodes, func(candidate *Node) bool {
		return candidate == node
	})
}
