package shape3d

import "github.com/mokiat/gomath/dprec"

// Rotation represents a 3D rotation encoded as a triplet of orthonormal basis
// vectors. BasisX, BasisY and BasisZ are the columns of the corresponding
// rotation matrix: each is the image of the respective axis under the rotation.
type Rotation struct {
	// BasisX is the direction the X axis maps to under this rotation.
	BasisX dprec.Vec3
	// BasisY is the direction the Y axis maps to under this rotation.
	BasisY dprec.Vec3
	// BasisZ is the direction the Z axis maps to under this rotation.
	BasisZ dprec.Vec3
}

// IdentityRotation returns a rotation that does not change the orientation of
// points it is applied to.
func IdentityRotation() Rotation {
	return Rotation{
		BasisX: dprec.BasisXVec3(),
		BasisY: dprec.BasisYVec3(),
		BasisZ: dprec.BasisZVec3(),
	}
}

// RotationFromQuat creates a Rotation corresponding to the given quaternion.
func RotationFromQuat(quat dprec.Quat) Rotation {
	return Rotation{
		BasisX: quat.OrientationX(),
		BasisY: quat.OrientationY(),
		BasisZ: quat.OrientationZ(),
	}
}

// ChainedRotation returns the composition of a parent rotation and a child
// rotation. The child is applied first and the parent second, so that
// ChainedRotation(parent, child).Apply(p) equals parent.Apply(child.Apply(p)).
func ChainedRotation(parent, child Rotation) Rotation {
	return Rotation{
		BasisX: parent.Apply(child.BasisX),
		BasisY: parent.Apply(child.BasisY),
		BasisZ: parent.Apply(child.BasisZ),
	}
}

// Quat returns the quaternion that represents this rotation.
func (r Rotation) Quat() dprec.Quat {
	mat := dprec.OrientationMat4(r.BasisX, r.BasisY, r.BasisZ)
	return mat.Rotation()
}

// Inverse returns the rotation that undoes this rotation. For an orthonormal
// rotation matrix this is equivalent to the transpose.
func (r Rotation) Inverse() Rotation {
	return Rotation{
		BasisX: dprec.NewVec3(r.BasisX.X, r.BasisY.X, r.BasisZ.X),
		BasisY: dprec.NewVec3(r.BasisX.Y, r.BasisY.Y, r.BasisZ.Y),
		BasisZ: dprec.NewVec3(r.BasisX.Z, r.BasisY.Z, r.BasisZ.Z),
	}
}

// Apply rotates the given point using this rotation.
func (r Rotation) Apply(point dprec.Vec3) dprec.Vec3 {
	return dprec.Vec3Sum(
		dprec.Vec3Prod(r.BasisX, point.X),
		dprec.Vec3Sum(
			dprec.Vec3Prod(r.BasisY, point.Y),
			dprec.Vec3Prod(r.BasisZ, point.Z),
		),
	)
}
