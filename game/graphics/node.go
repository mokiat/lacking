package graphics

import "github.com/mokiat/gomath/sprec"

func newNode() *Node {
	return &Node{
		position: sprec.ZeroVec3(),
		rotation: sprec.IdentityQuat(),
		scale:    sprec.NewVec3(1.0, 1.0, 1.0),
	}
}

// Node represents a positioning of some entity in
// the 3D scene.
type Node struct {
	position sprec.Vec3
	rotation sprec.Quat
	scale    sprec.Vec3
}

// Position returns this entity's position.
func (n *Node) Position() sprec.Vec3 {
	return n.position
}

// SetPosition changes this entity's position.
func (n *Node) SetPosition(position sprec.Vec3) {
	n.position = position
}

// Rotation returns this entity's rotation.
func (n *Node) Rotation() sprec.Quat {
	return n.rotation
}

// SetRotation changes this entity's rotation.
func (n *Node) SetRotation(rotation sprec.Quat) {
	n.rotation = sprec.UnitQuat(rotation)
}

// Scale returns this entity's scale.
func (n *Node) Scale() sprec.Vec3 {
	return n.scale
}

// SetScale changes this entity's scale.
func (n *Node) SetScale(scale sprec.Vec3) {
	n.scale = scale
}

// Matrix returns the matrix transformation
// of this node.
func (n *Node) Matrix() sprec.Mat4 {
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
