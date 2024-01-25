package solver

import "github.com/mokiat/gomath/dprec"

// Medium represents a medium that can be used to simulate
// the effects of drag and lift.
type Medium interface {

	// Density returns the density of the medium at the specified
	// position.
	Density(position dprec.Vec3) float64

	// Velocity returns the velocity of the medium at the specified
	// position.
	Velocity(position dprec.Vec3) dprec.Vec3
}
