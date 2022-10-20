package shape

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
)

// IdentityTransform returns a new Transform that represents the origin.
func IdentityTransform() Transform {
	return Transform{
		position: dprec.ZeroVec3(),
		rotation: dprec.IdentityQuat(),
	}
}

// NewTransform creates a new Transform with the specified position and
// rotation.
func NewTransform(position dprec.Vec3, rotation dprec.Quat) Transform {
	return Transform{
		position: position,
		rotation: rotation,
	}
}

// Transform represents a shape transformation - translation and rotation.
type Transform struct {
	position dprec.Vec3
	rotation dprec.Quat
}

func (t Transform) IsIdentity() bool {
	return t.position.IsZero() && t.rotation.IsIdentity()
}

// Position returns the translation of this Transform.
func (t Transform) Position() dprec.Vec3 {
	return t.position
}

// Rotation returns the orientation of this Transform.
func (t Transform) Rotation() dprec.Quat {
	return t.rotation
}

// Transformed returns a new Transform that is based on this one but has the
// specified Transform applied to it.
func (t Transform) Transformed(transform Transform) Transform {
	if transform.IsIdentity() {
		return t
	}
	return Transform{
		position: dprec.Vec3Sum(
			transform.position,
			dprec.QuatVec3Rotation(transform.rotation, t.position),
		),
		rotation: dprec.QuatProd(transform.rotation, t.rotation),
	}
}

// TransformedShape is a helper method that transforms the specified Shape
// using the specified Transform.
//
// This is a helper method necessary due to Go's generics limitations at
// the moment.
func TransformedShape(sh Shape, transform Transform) Shape {
	if transform.IsIdentity() {
		return sh
	}
	switch sh := sh.(type) {
	case StaticSphere:
		return sh.Transformed(transform)
	case StaticBox:
		return sh.Transformed(transform)
	case StaticMesh:
		return sh.Transformed(transform)
	default:
		panic(fmt.Errorf("unknown shape type %T", sh))
	}
}
