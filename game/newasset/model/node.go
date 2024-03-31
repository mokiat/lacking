package model

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

var _ Node = (*BlankNode)(nil)

type BlankNode struct {
	name string

	translation dprec.Vec3
	rotation    dprec.Quat
	scale       dprec.Vec3

	parent Node
	nodes  []Node
}

func (n *BlankNode) Name() string {
	return n.name
}

func (n *BlankNode) SetName(name string) {
	n.name = name
}

func (n *BlankNode) Translation() dprec.Vec3 {
	return n.translation
}

func (n *BlankNode) SetTranslation(translation dprec.Vec3) {
	n.translation = translation
}

func (n *BlankNode) Rotation() dprec.Quat {
	return n.rotation
}

func (n *BlankNode) SetRotation(rotation dprec.Quat) {
	n.rotation = rotation
}

func (n *BlankNode) Scale() dprec.Vec3 {
	return n.scale
}

func (n *BlankNode) SetScale(scale dprec.Vec3) {
	n.scale = scale
}

func (n *BlankNode) Parent() Node {
	return n.parent
}

func (n *BlankNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *BlankNode) Nodes() []Node {
	return n.nodes
}

func (n *BlankNode) AddNode(node Node) {
	n.nodes = append(n.nodes, node)
}

func (n *BlankNode) RemoveNode(node Node) {
	n.nodes = slices.DeleteFunc(n.nodes, func(candidate Node) bool {
		return candidate == node
	})
}
