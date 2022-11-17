package physics

import "github.com/mokiat/gomath/dprec"

// SymmetricMomentOfInertia returns a moment of inertia
// tensor that represents a symmetric object across all
// axis.
func SymmetricMomentOfInertia(value float64) dprec.Mat3 {
	return dprec.NewMat3(
		value, 0.0, 0.0,
		0.0, value, 0.0,
		0.0, 0.0, value,
	)
}

// SolidSphereMomentOfInertia returns the moment of inertia of a solid
// sphere with the specified mass and radius.
func SolidSphereMomentOfInertia(mass, radius float64) dprec.Mat3 {
	return SymmetricMomentOfInertia(mass * radius * radius * (2.0 / 5.0))
}

// HollowSphereMomentOfInertia returns the moment of inertia of a hollow
// sphere with the specified mass and radius.
func HollowSphereMomentOfInertia(mass, radius float64) dprec.Mat3 {
	return SymmetricMomentOfInertia(mass * radius * radius * (2.0 / 3.0))
}
