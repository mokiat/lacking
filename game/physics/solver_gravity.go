package physics

import "github.com/mokiat/gomath/dprec"

var _ AccelerationSolver = (*GravitySolver)(nil)

// GravitySolver is an [AccelerationSolver] that applies a constant
// acceleration along a fixed direction.
//
// Since it contributes an acceleration and not a force, it pulls on all
// bodies equally, regardless of their mass.
type GravitySolver struct {
	direction dprec.Vec3
	magnitude float64

	acceleration dprec.Vec3
}

// NewGravitySolver creates a new [GravitySolver] that is configured with
// the gravity of Earth, pulling downward along the negative Y axis.
func NewGravitySolver() *GravitySolver {
	result := &GravitySolver{
		direction: dprec.NewVec3(0.0, -1.0, 0.0),
		magnitude: 9.8,
	}
	result.refreshAcceleration()
	return result
}

// Direction returns the direction along which gravity pulls, as a unit
// vector in world space.
func (s *GravitySolver) Direction() dprec.Vec3 {
	return s.direction
}

// SetDirection changes the direction along which gravity pulls. The
// specified direction need not be normalized.
//
// It returns the solver itself, so that calls can be chained.
func (s *GravitySolver) SetDirection(direction dprec.Vec3) *GravitySolver {
	s.direction = dprec.UnitVec3(direction)
	s.refreshAcceleration()
	return s
}

// Magnitude returns the magnitude of the gravitational acceleration,
// in m/s^2.
func (s *GravitySolver) Magnitude() float64 {
	return s.magnitude
}

// SetMagnitude changes the magnitude of the gravitational acceleration,
// in m/s^2.
//
// It returns the solver itself, so that calls can be chained.
func (s *GravitySolver) SetMagnitude(magnitude float64) *GravitySolver {
	s.magnitude = magnitude
	s.refreshAcceleration()
	return s
}

// ApplyAcceleration accumulates the gravitational acceleration on the
// target. The medium of the context is not taken into account, meaning that
// buoyancy is not modeled.
func (s *GravitySolver) ApplyAcceleration(ctx AccelerationContext, target *AccelerationTarget) {
	target.AddLinearAcceleration(s.acceleration)
}

func (s *GravitySolver) refreshAcceleration() {
	s.acceleration = dprec.Vec3Prod(s.direction, s.magnitude)
}
