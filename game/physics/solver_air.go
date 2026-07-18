package physics

import "github.com/mokiat/gomath/dprec"

var _ MediumSolver = (*StaticAirSolver)(nil)

// StaticAirSolver is a [MediumSolver] that models air with a uniform density
// and a uniform velocity throughout the whole scene.
//
// This is the default medium solver of a [Scene] and is a good fit for
// simulations that do not need altitude-dependent air or localized wind.
type StaticAirSolver struct {
	airDensity  float64
	airVelocity dprec.Vec3
}

// NewStaticAirSolver creates a new [StaticAirSolver] that is configured with
// the density of air at sea level and with no wind.
func NewStaticAirSolver() *StaticAirSolver {
	return &StaticAirSolver{
		airDensity:  1.2,
		airVelocity: dprec.ZeroVec3(),
	}
}

// AirDensity returns the density of the air, in kg/m^3.
func (s *StaticAirSolver) AirDensity() float64 {
	return s.airDensity
}

// SetAirDensity changes the density of the air, in kg/m^3.
//
// It returns the solver itself, so that calls can be chained.
func (s *StaticAirSolver) SetAirDensity(density float64) *StaticAirSolver {
	s.airDensity = density
	return s
}

// AirVelocity returns the velocity of the air in world space, in m/s. This
// is essentially the wind that is applied to the whole scene.
func (s *StaticAirSolver) AirVelocity() dprec.Vec3 {
	return s.airVelocity
}

// SetAirVelocity changes the velocity of the air in world space, in m/s.
//
// It returns the solver itself, so that calls can be chained.
func (s *StaticAirSolver) SetAirVelocity(velocity dprec.Vec3) *StaticAirSolver {
	s.airVelocity = velocity
	return s
}

// Density returns the density of the air at the specified world position.
//
// The position is ignored, since the density is uniform.
func (s *StaticAirSolver) Density(position dprec.Vec3) float64 {
	return s.airDensity
}

// Velocity returns the velocity of the air at the specified world position.
//
// The position is ignored, since the velocity is uniform.
func (s *StaticAirSolver) Velocity(position dprec.Vec3) dprec.Vec3 {
	return s.airVelocity
}
