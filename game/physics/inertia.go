package physics

import "github.com/mokiat/gomath/sprec"

// SymmetricMomentOfInertia returns a moment of inertia
// tensor that represents a symmetric object across all
// axis.
func SymmetricMomentOfInertia(value float32) sprec.Mat3 {
	return sprec.NewMat3(
		value, 0.0, 0.0,
		0.0, value, 0.0,
		0.0, 0.0, value,
	)
}
