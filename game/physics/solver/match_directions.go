package solver

import (
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewMatchDirections creates a new MatchDirections constraint solver.
func NewMatchDirections() *MatchDirections {
	return &MatchDirections{
		primaryDirection:   dprec.BasisYVec3(),
		secondaryDirection: dprec.BasisYVec3(),
	}
}

var _ physics.ExplicitDBConstraintSolver = (*MatchDirections)(nil)

// MatchDirections represents the solution for a constraint
// that keeps the direction of two bodies pointing in the same
// direction.
type MatchDirections struct {
	physics.NilDBConstraintSolver // TODO: Remove

	primaryDirection   dprec.Vec3
	secondaryDirection dprec.Vec3

	jacobian physics.PairJacobian
	drift    float64
}

// PrimaryDirection returns the direction of the primary body that will be
// used in the alignment.
func (s *MatchDirections) PrimaryDirection() dprec.Vec3 {
	return s.primaryDirection
}

// SetPrimaryDirection changes the direction of the primary body to be used
// in the alignment.
func (s *MatchDirections) SetPrimaryDirection(direction dprec.Vec3) *MatchDirections {
	s.primaryDirection = dprec.UnitVec3(direction)
	return s
}

// SecondaryDirection returns the direction of the secondary body that will be
// used in the alignment.
func (s *MatchDirections) SecondaryDirection() dprec.Vec3 {
	return s.secondaryDirection
}

// SetSecondaryDirection changes the direction of the secondary body to be
// used in the alignment.
func (s *MatchDirections) SetSecondaryDirection(direction dprec.Vec3) *MatchDirections {
	s.secondaryDirection = dprec.UnitVec3(direction)
	return s
}

func (s *MatchDirections) Reset(ctx physics.DBSolverContext) {
	s.updateJacobian(ctx)
}

func (s *MatchDirections) ApplyImpulses(ctx physics.DBSolverContext) {
	if dprec.Abs(s.drift) > epsilon {
		ctx.ApplyImpulse(s.jacobian)
	}
}

func (s *MatchDirections) ApplyNudges(ctx physics.DBSolverContext) {
	s.updateJacobian(ctx)
	if dprec.Abs(s.drift) > epsilon {
		ctx.ApplyNudge(s.jacobian, s.drift)
	}
}

func (s *MatchDirections) updateJacobian(ctx physics.DBSolverContext) {
	primaryDirWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.primaryDirection)
	secondaryDirWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), s.secondaryDirection)
	cross := dprec.Vec3Cross(secondaryDirWS, primaryDirWS)
	slope := SafeNormal(cross, dprec.BasisYVec3())

	s.jacobian = physics.PairJacobian{
		Primary: physics.Jacobian{
			SlopeVelocity:        dprec.ZeroVec3(),
			SlopeAngularVelocity: slope,
		},
		Secondary: physics.Jacobian{
			SlopeVelocity:        dprec.ZeroVec3(),
			SlopeAngularVelocity: dprec.InverseVec3(slope),
		},
	}

	cos := dprec.Vec3Dot(secondaryDirWS, primaryDirWS)
	sin := cross.Length()
	s.drift = math.Atan2(sin, cos)
}
