package shape2d

import "github.com/mokiat/gomath/sprec"

// Rotation represents a 2D rotation encoded as a pair of orthonormal basis
// vectors. BasisX and BasisY are the columns of the corresponding rotation
// matrix: BasisX is the image of the X axis and BasisY is the image of the Y
// axis under the rotation.
type Rotation struct {
	// BasisX is the direction the X axis maps to under this rotation.
	BasisX sprec.Vec2
	// BasisY is the direction the Y axis maps to under this rotation.
	BasisY sprec.Vec2
}

// IdentityRotation returns a rotation that does not change the orientation of
// points it is applied to.
func IdentityRotation() Rotation {
	return Rotation{
		BasisX: sprec.BasisXVec2(),
		BasisY: sprec.BasisYVec2(),
	}
}

// RotationFromAngle creates a Rotation corresponding to the given angle.
// Positive angles rotate counter-clockwise.
func RotationFromAngle(angle sprec.Angle) Rotation {
	cos := sprec.Cos(angle)
	sin := sprec.Sin(angle)
	return RotationFromCosSin(cos, sin)
}

// RotationFromCosSin creates a Rotation from precomputed cosine and sine
// values of the rotation angle.
func RotationFromCosSin(cos, sin float32) Rotation {
	return Rotation{
		BasisX: sprec.NewVec2(cos, sin),
		BasisY: sprec.NewVec2(-sin, cos),
	}
}

// Angle returns the rotation angle in the range [-Pi, Pi].
func (r Rotation) Angle() sprec.Angle {
	return sprec.Atan2(r.BasisX.Y, r.BasisX.X)
}

// Inverse returns the rotation that undoes this rotation. For an orthonormal
// rotation matrix this is equivalent to the transpose.
func (r Rotation) Inverse() Rotation {
	return Rotation{
		BasisX: sprec.NewVec2(r.BasisX.X, r.BasisY.X),
		BasisY: sprec.NewVec2(r.BasisX.Y, r.BasisY.Y),
	}
}

// Apply rotates the given point using this rotation.
func (r Rotation) Apply(point sprec.Vec2) sprec.Vec2 {
	return sprec.Vec2Sum(
		sprec.Vec2Prod(r.BasisX, point.X),
		sprec.Vec2Prod(r.BasisY, point.Y),
	)
}
