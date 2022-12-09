package solver

import "github.com/mokiat/gomath/dprec"

// Epsilon indicates a small enough amount that something could be ignored.
const Epsilon = float64(0.00001)

// RestitutionClamp specifies a ratio that describes how much the restitution
// coefficient should be allowed to apply.
//
// The goal of this clamp is to reduce bounciness of objects when they are
// barely moving.
func RestitutionClamp(effectiveVelocity float64) float64 {
	absEffectiveVelocity := dprec.Abs(effectiveVelocity)
	switch {
	case absEffectiveVelocity < 2.0:
		return 0.1
	case absEffectiveVelocity < 1.0:
		return 0.05
	case absEffectiveVelocity < 0.5:
		return 0.0
	default:
		return 1.0
	}
}
