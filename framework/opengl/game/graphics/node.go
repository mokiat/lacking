package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
)

func newNode() Node {
	return Node{
		position: sprec.ZeroVec3(),
		rotation: sprec.IdentityQuat(),
		scale:    sprec.NewVec3(1.0, 1.0, 1.0),
	}
}

var _ graphics.Node = (*Node)(nil)

type Node struct {
	position sprec.Vec3
	rotation sprec.Quat
	scale    sprec.Vec3
}

func (n *Node) Position() sprec.Vec3 {
	return n.position
}

func (n *Node) SetPosition(position sprec.Vec3) {
	n.position = position
}

func (n *Node) Rotation() sprec.Quat {
	return n.rotation
}

func (n *Node) SetRotation(rotation sprec.Quat) {
	n.rotation = sprec.UnitQuat(rotation)
}

func (n *Node) Scale() sprec.Vec3 {
	return n.scale
}
func (n *Node) SetScale(scale sprec.Vec3) {
	n.scale = scale
}
