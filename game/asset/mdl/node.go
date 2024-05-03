package mdl

import (
	"slices"

	"github.com/mokiat/gomath/dprec"
)

type NodeContainer interface {
	AddNode(node Node)
	RemoveNode(node Node)
	Nodes() []Node
}

type Node interface {
	Name() string
	SetName(name string)

	Translatable
	Rotatable
	Scalable

	Parent() Node
	SetParent(parent Node)

	NodeContainer
}

type Translatable interface {
	Translation() dprec.Vec3
	SetTranslation(translation dprec.Vec3)
}

type Rotatable interface {
	Rotation() dprec.Quat
	SetRotation(rotation dprec.Quat)
}

type Scalable interface {
	Scale() dprec.Vec3
	SetScale(scale dprec.Vec3)
}

var _ Node = (*BaseNode)(nil)

type BaseNode struct {
	name string

	translation dprec.Vec3
	rotation    dprec.Quat
	scale       dprec.Vec3

	parent Node
	nodes  []Node
}

func (n *BaseNode) Name() string {
	return n.name
}

func (n *BaseNode) SetName(name string) {
	n.name = name
}

func (n *BaseNode) Translation() dprec.Vec3 {
	return n.translation
}

func (n *BaseNode) SetTranslation(translation dprec.Vec3) {
	n.translation = translation
}

func (n *BaseNode) Rotation() dprec.Quat {
	return n.rotation
}

func (n *BaseNode) SetRotation(rotation dprec.Quat) {
	n.rotation = rotation
}

func (n *BaseNode) Scale() dprec.Vec3 {
	return n.scale
}

func (n *BaseNode) SetScale(scale dprec.Vec3) {
	n.scale = scale
}

func (n *BaseNode) Parent() Node {
	return n.parent
}

func (n *BaseNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BaseNode) Nodes() []Node {
	return n.nodes
}

func (n *BaseNode) AddNode(node Node) {
	node.SetParent(n)
	n.nodes = append(n.nodes, node)
}

func (n *BaseNode) RemoveNode(node Node) {
	if node.Parent() == n {
		node.SetParent(nil)
	}
	n.nodes = slices.DeleteFunc(n.nodes, func(candidate Node) bool {
		return candidate == node
	})
}
