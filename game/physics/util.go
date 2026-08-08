package physics

import "github.com/mokiat/gomath/dprec"

// Epsilon is a threshold below which a quantity is small enough to be
// treated as zero, in order to avoid degenerate behavior such as
// normalizing a near-zero-length vector.
const Epsilon = float64(0.00001)

// QuatFromVector returns the quaternion that represents a rotation of
// vector's length, in radians, around vector's direction as the rotation
// axis.
//
// This is the standard way to turn a rotation vector (e.g. an angular
// velocity scaled by elapsed time, or an angular nudge) into a quaternion
// that can be composed with an existing orientation.
//
// If vector's length is smaller than [Epsilon], the identity quaternion
// is returned instead, since the direction of a near-zero vector is not
// meaningful as a rotation axis.
func QuatFromVector(vector dprec.Vec3) dprec.Quat {
	radians := vector.Length()
	if dprec.Abs(radians) < Epsilon {
		return dprec.IdentityQuat()
	}
	return dprec.RotationQuat(dprec.Radians(radians), vector)
}

// CombinedFrictionCoefficient returns the friction coefficient to use for a
// contact between two materials whose individual friction coefficients are
// first and second, combining them as their geometric mean.
//
// This is independent of which kind of solver the contact belongs to; it
// applies equally to a body-body contact (e.g. [PairCollisionSolver]) and
// a body-terrain contact (e.g. [SoloCollisionSolver]).
func CombinedFrictionCoefficient(first, second float64) float64 {
	return dprec.Sqrt(first * second)
}

// CombinedRestitutionCoefficient returns the restitution coefficient to use
// for a contact between two materials whose individual restitution
// coefficients are first and second, combining them by taking the larger
// of the two.
//
// This is independent of which kind of solver the contact belongs to; it
// applies equally to a body-body contact (e.g. [PairCollisionSolver]) and
// a body-terrain contact (e.g. [SoloCollisionSolver]).
func CombinedRestitutionCoefficient(first, second float64) float64 {
	return max(first, second)
}

// RestitutionClamp specifies a ratio that describes how much the restitution
// coefficient should be allowed to apply.
//
// The goal of this clamp is to reduce bounciness of objects when they are
// barely moving.
func RestitutionClamp(effectiveVelocity float64) float64 {
	absEffectiveVelocity := dprec.Abs(effectiveVelocity)
	switch {
	case absEffectiveVelocity < 0.5:
		return 0.0
	case absEffectiveVelocity < 1.0:
		return 0.05
	case absEffectiveVelocity < 2.0:
		return 0.1
	default:
		return 1.0
	}
}
