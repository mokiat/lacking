package model

import (
	"github.com/mokiat/gomath/dprec"
	asset "github.com/mokiat/lacking/game/newasset"
)

type NodeContainer interface {
	AddNode(node *Node)
}

type Node struct {
	name string
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) SetName(name string) {
	n.name = name
}

func (n *Node) ToAsset() (asset.Node, error) {
	result := asset.Node{
		Name:        n.name,
		ParentIndex: asset.UnspecifiedNodeIndex,
		Translation: dprec.ZeroVec3(),
		Rotation:    dprec.IdentityQuat(),
		Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		Mask:        asset.NodeMaskNone,
	}
	return result, nil
}
