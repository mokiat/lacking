package graphics

import "github.com/mokiat/gomath/sprec"

// Node represents a positioning of some entity in
// the 3D scene.
type Node interface {

	// Position returns this entity's position.
	Position() sprec.Vec3

	// SetPosition changes this entity's position.
	SetPosition(position sprec.Vec3)

	// Rotation returns this entity's rotation.
	Rotation() sprec.Quat

	// SetRotation changes this entity's rotation.
	SetRotation(rotation sprec.Quat)

	// Scale returns this entity's scale.
	Scale() sprec.Vec3

	// SetScale changes this entity's scale.
	SetScale(scale sprec.Vec3)
}
