package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewClampDirection creates a new ClampDirection constraint solver.
func NewClampDirection() *ClampDirection {
	return &ClampDirection{
		direction:   dprec.BasisYVec3(),
		min:         -1.0,
		max:         1.0,
		restitution: 0.0,
	}
}

var _ physics.ExplicitDBConstraintSolver = (*ClampDirection)(nil)

// ClampDirection represents the solution for a constraint which ensures that
// the second body is within certain min and max bounds relative to the first
// body along a certain direction of the first body.
type ClampDirection struct {
	physics.NilDBConstraintSolver // TODO: Remove

	direction   dprec.Vec3
	min         float64
	max         float64
	restitution float64

	jacobian physics.PairJacobian
	drift    float64
}

// Direction returns the constraint direction, which is in local space of
// the first body.
func (s *ClampDirection) Direction() dprec.Vec3 {
	return s.direction
}

// SetDirection changes the constraint direction, which must be in local space
// of the first body.
func (s *ClampDirection) SetDirection(direction dprec.Vec3) *ClampDirection {
	s.direction = dprec.UnitVec3(direction)
	return s
}

// Min returns the lower bounds limit.
func (s *ClampDirection) Min() float64 {
	return s.min
}

// SetMin changes the lower bounds limit.
func (s *ClampDirection) SetMin(min float64) *ClampDirection {
	s.min = min
	return s
}

// Max returns the upper bounds limit.
func (s *ClampDirection) Max() float64 {
	return s.max
}

// SetMax changes the upper bounds limit.
func (s *ClampDirection) SetMax(max float64) *ClampDirection {
	s.max = max
	return s
}

// Restitution returns the restitution to be used when adjusting the
// two bodies when the constraint is not met.
func (s *ClampDirection) Restitution() float64 {
	return s.restitution
}

// SetRestitution changes the restitution to be used when adjusting the
// two bodies when the constraint is not met.
func (s *ClampDirection) SetRestitution(restitution float64) *ClampDirection {
	s.restitution = restitution
	return s
}

func (s *ClampDirection) Reset(ctx physics.DBSolverContext) {
	s.updateJacobian(ctx)
}

func (s *ClampDirection) ApplyImpulses(ctx physics.DBSolverContext) {
	if s.drift > 0.0 {
		ctx.ApplyElasticImpulse(s.jacobian, s.restitution)
	}
}

func (s *ClampDirection) ApplyNudges(ctx physics.DBSolverContext) {
	s.updateJacobian(ctx)
	if s.drift > 0.0 {
		ctx.ApplyNudge(s.jacobian, s.drift)
	}
}

func (s *ClampDirection) updateJacobian(ctx physics.DBSolverContext) {
	dirWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.direction)
	deltaPosition := dprec.Vec3Diff(ctx.Secondary.Position(), ctx.Primary.Position())
	dirDistance := dprec.Vec3Dot(deltaPosition, dirWS)

	switch {
	case dirDistance > s.max:
		radius := dprec.Vec3Diff(
			deltaPosition,
			dprec.Vec3Prod(dirWS, dirDistance-s.max),
		)
		s.jacobian = physics.PairJacobian{
			Primary: physics.Jacobian{
				SlopeVelocity:        dprec.InverseVec3(dirWS),
				SlopeAngularVelocity: dprec.Vec3Cross(dirWS, radius),
			},
			Secondary: physics.Jacobian{
				SlopeVelocity:        dirWS,
				SlopeAngularVelocity: dprec.ZeroVec3(),
			},
		}
		s.drift = dirDistance - s.max

	case dirDistance < s.min:
		radius := dprec.Vec3Sum(
			deltaPosition,
			dprec.Vec3Prod(dirWS, s.min-dirDistance),
		)
		s.jacobian = physics.PairJacobian{
			Primary: physics.Jacobian{
				SlopeVelocity:        dirWS,
				SlopeAngularVelocity: dprec.Vec3Cross(radius, dirWS),
			},
			Secondary: physics.Jacobian{
				SlopeVelocity:        dprec.InverseVec3(dirWS),
				SlopeAngularVelocity: dprec.ZeroVec3(),
			},
		}
		s.drift = s.min - dirDistance

	default:
		s.jacobian = physics.PairJacobian{}
		s.drift = 0
	}
}
