package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
)

func NewNode() Node {
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

func (n *Node) ModelMatrix() sprec.Mat4 {
	return sprec.Mat4MultiProd(
		sprec.TranslationMat4(
			n.position.X,
			n.position.Y,
			n.position.Z,
		),
		sprec.OrientationMat4(
			n.rotation.OrientationX(),
			n.rotation.OrientationY(),
			n.rotation.OrientationZ(),
		),
		sprec.ScaleMat4(
			n.scale.X,
			n.scale.Y,
			n.scale.Z,
		),
	)
}
