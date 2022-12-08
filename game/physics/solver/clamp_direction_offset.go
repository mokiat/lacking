package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewClampDirectionOffset creates a new ClampDirectionOffset constraint solver.
func NewClampDirectionOffset() *ClampDirectionOffset {
	return &ClampDirectionOffset{
		direction:   dprec.BasisYVec3(),
		min:         -1.0,
		max:         1.0,
		restitution: 0.0,
	}
}

var _ physics.DBConstraintSolver = (*ClampDirectionOffset)(nil)

// ClampDirectionOffset represents the solution for a constraint which ensures that
// the second body is within certain min and max bounds relative to the first
// body along a certain direction of the first body.
type ClampDirectionOffset struct {
	direction   dprec.Vec3
	min         float64
	max         float64
	restitution float64

	jacobian physics.PairJacobian
	drift    float64
}

// Direction returns the constraint direction, which is in local space of
// the first body.
func (s *ClampDirectionOffset) Direction() dprec.Vec3 {
	return s.direction
}

// SetDirection changes the constraint direction, which must be in local space
// of the first body.
func (s *ClampDirectionOffset) SetDirection(direction dprec.Vec3) *ClampDirectionOffset {
	s.direction = dprec.UnitVec3(direction)
	return s
}

// Min returns the lower bounds limit.
func (s *ClampDirectionOffset) Min() float64 {
	return s.min
}

// SetMin changes the lower bounds limit.
func (s *ClampDirectionOffset) SetMin(min float64) *ClampDirectionOffset {
	s.min = min
	return s
}

// Max returns the upper bounds limit.
func (s *ClampDirectionOffset) Max() float64 {
	return s.max
}

// SetMax changes the upper bounds limit.
func (s *ClampDirectionOffset) SetMax(max float64) *ClampDirectionOffset {
	s.max = max
	return s
}

// Restitution returns the restitution to be used when adjusting the
// two bodies when the constraint is not met.
func (s *ClampDirectionOffset) Restitution() float64 {
	return s.restitution
}

// SetRestitution changes the restitution to be used when adjusting the
// two bodies when the constraint is not met.
func (s *ClampDirectionOffset) SetRestitution(restitution float64) *ClampDirectionOffset {
	s.restitution = restitution
	return s
}

func (s *ClampDirectionOffset) Reset(ctx physics.DBSolverContext) {
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

func (s *ClampDirectionOffset) ApplyImpulses(ctx physics.DBSolverContext) {
	lambda := ctx.JacobianImpulseLambda(s.jacobian, s.drift, s.restitution)
	if lambda > 0.0 {
		return // moving away
	}
	solution := s.jacobian.ImpulseSolution(lambda)
	ctx.ApplyImpulseSolution(solution)
}

func (s *ClampDirectionOffset) ApplyNudges(ctx physics.DBSolverContext) {
	if s.drift > 0 {
		solution := ctx.JacobianNudgeSolution(s.jacobian, s.drift)
		ctx.ApplyNudgeSolution(solution)
	}
}
