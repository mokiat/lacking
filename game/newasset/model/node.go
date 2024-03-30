package model

import (
	"github.com/mokiat/gomath/dprec"
	asset "github.com/mokiat/lacking/game/newasset"
)

type NodeContainer interface {
	AddNode(node *Node)
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

type ContentHolder interface {
	Content() any
	SetContent(content any)
}

type Node struct {
	name        string
	translation dprec.Vec3
	rotation    dprec.Quat
	scale       dprec.Vec3
	content     any
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

func (n *Node) Content() any {
	return n.content
}

func (n *Node) SetContent(content any) {
	n.content = content
}

func (n *Node) ToAsset() (asset.Node, error) {
	result := asset.Node{
		Name:        n.name,
		ParentIndex: asset.UnspecifiedNodeIndex,
		Translation: n.translation,
		Rotation:    n.rotation,
		Scale:       n.scale,
		Mask:        asset.NodeMaskNone,
	}
	return result, nil
}
