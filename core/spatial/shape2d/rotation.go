package shape2d

import "github.com/mokiat/gomath/dprec"

// Rotation represents a 2D rotation encoded as a pair of orthonormal basis
// vectors. BasisX and BasisY are the columns of the corresponding rotation
// matrix: BasisX is the image of the X axis and BasisY is the image of the Y
// axis under the rotation.
type Rotation struct {
	// BasisX is the direction the X axis maps to under this rotation.
	BasisX dprec.Vec2
	// BasisY is the direction the Y axis maps to under this rotation.
	BasisY dprec.Vec2
}

// IdentityRotation returns a rotation that does not change the orientation of
// points it is applied to.
func IdentityRotation() Rotation {
	return Rotation{
		BasisX: dprec.BasisXVec2(),
		BasisY: dprec.BasisYVec2(),
	}
}

// RotationFromAngle creates a Rotation corresponding to the given angle.
// Positive angles rotate counter-clockwise.
func RotationFromAngle(angle dprec.Angle) Rotation {
	cos := dprec.Cos(angle)
	sin := dprec.Sin(angle)
	return RotationFromCosSin(cos, sin)
}

// RotationFromCosSin creates a Rotation from precomputed cosine and sine
// values of the rotation angle.
func RotationFromCosSin(cos, sin float64) Rotation {
	return Rotation{
		BasisX: dprec.NewVec2(cos, sin),
		BasisY: dprec.NewVec2(-sin, cos),
	}
}

// ChainedRotation returns the composition of a parent rotation and a child
// rotation. The child is applied first and the parent second, so that
// ChainedRotation(parent, child).Apply(p) equals parent.Apply(child.Apply(p)).
func ChainedRotation(parent, child Rotation) Rotation {
	return Rotation{
		BasisX: parent.Apply(child.BasisX),
		BasisY: parent.Apply(child.BasisY),
	}
}

// Angle returns the rotation angle in the range [-Pi, Pi].
func (r Rotation) Angle() dprec.Angle {
	return dprec.Atan2(r.BasisX.Y, r.BasisX.X)
}

// Inverse returns the rotation that undoes this rotation. For an orthonormal
// rotation matrix this is equivalent to the transpose.
func (r Rotation) Inverse() Rotation {
	return Rotation{
		BasisX: dprec.NewVec2(r.BasisX.X, r.BasisY.X),
		BasisY: dprec.NewVec2(r.BasisX.Y, r.BasisY.Y),
	}
}

// Apply rotates the given point using this rotation.
func (r Rotation) Apply(point dprec.Vec2) dprec.Vec2 {
	return dprec.Vec2Sum(
		dprec.Vec2Prod(r.BasisX, point.X),
		dprec.Vec2Prod(r.BasisY, point.Y),
	)
}
