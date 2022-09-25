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
