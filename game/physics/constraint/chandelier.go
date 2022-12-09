package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewChandelier creates a new Chandelier constraint solver.
func NewChandelier() *Chandelier {
	return &Chandelier{
		fixture: dprec.ZeroVec3(),
		radius:  dprec.ZeroVec3(),
		length:  1.0,
	}
}

var _ solver.Constraint = (*Chandelier)(nil)

// Chandelier represents the solution for a constraint
// that keeps a body hanging off of a fixture location similar
// to a chandelier.
type Chandelier struct {
	fixture dprec.Vec3
	radius  dprec.Vec3
	length  float64

	jacobian solver.Jacobian
	drift    float64
}

// Fixture returns the fixture location for the chandelier hook.
func (s *Chandelier) Fixture() dprec.Vec3 {
	return s.fixture
}

// SetFixture changes the fixture location for the chandelier hook.
func (s *Chandelier) SetFixture(fixture dprec.Vec3) *Chandelier {
	s.fixture = fixture
	return s
}

// Radius returns the radius vector of the contact point on the object.
//
// The vector is in the object's local space.
func (s *Chandelier) Radius() dprec.Vec3 {
	return s.radius
}

// SetRadius changes the radius vector of the contact point on the object.
//
// The vector is in the object's local space.
func (s *Chandelier) SetRadius(radius dprec.Vec3) *Chandelier {
	s.radius = radius
	return s
}

// Length returns the chandelier length.
func (s *Chandelier) Length() float64 {
	return s.length
}

// SetLength changes the chandelier length.
func (s *Chandelier) SetLength(length float64) *Chandelier {
	s.length = length
	return s
}

// Reset re-evaluates the constraint.
func (s *Chandelier) Reset(ctx solver.Context) {
	radiusWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.radius)
	pointWS := dprec.Vec3Sum(ctx.Target.Position(), radiusWS)
	deltaPositionWS := dprec.Vec3Diff(pointWS, s.fixture)
	if distance := deltaPositionWS.Length(); distance > solver.Epsilon {
		normalWS := dprec.Vec3Quot(deltaPositionWS, distance)
		s.jacobian = solver.Jacobian{
			LinearSlope:  normalWS,
			AngularSlope: dprec.Vec3Cross(radiusWS, normalWS),
		}
		s.drift = distance - s.length
	} else {
		s.jacobian = solver.Jacobian{}
		s.drift = -s.length
	}
}

// ApplyImpulses applies impulses in order to keep the velocity part of
// the constraint satisfied.
func (s *Chandelier) ApplyImpulses(ctx solver.Context) {
	solution := ctx.JacobianImpulseSolution(s.jacobian, s.drift, 0.0)
	ctx.Target.ApplyImpulse(solution)
}

// ApplyNudges applies nudges in order to keep the positional part of the
// constraint satisfied.
func (s *Chandelier) ApplyNudges(ctx solver.Context) {
	solution := ctx.JacobianNudgeSolution(s.jacobian, s.drift)
	ctx.Target.ApplyNudge(solution)
}
