package physics

import "github.com/mokiat/gomath/dprec"

// MediumSolver describes the medium (e.g. air or water) that surrounds the
// bodies of a [Scene].
//
// It is sampled once per body per simulation step and the result is passed
// to the acceleration contributors through an [AccelerationContext], which
// is what allows effects like drag and lift to be evaluated.
//
// Implementations must be safe to call repeatedly within a single step and
// must not mutate scene state.
type MediumSolver interface {

	// Velocity returns the velocity of the medium at the specified world
	// position, in m/s.
	Velocity(position dprec.Vec3) dprec.Vec3

	// Density returns the density of the medium at the specified world
	// position, in kg/m^3.
	Density(position dprec.Vec3) float64
}
