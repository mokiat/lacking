package solver

import (
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

var _ physics.DBConstraintSolver = (*MatchDirections)(nil)

// MatchDirections represents the solution for a constraint
// that keeps the direction of two bodies pointing in the same
// direction.
type MatchDirections struct {
	primaryDirection   dprec.Vec3
	secondaryDirection dprec.Vec3

	jacobian1 physics.PairJacobian
	jacobian2 physics.PairJacobian
	drift1    float64
	drift2    float64
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
	primaryDirWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.primaryDirection)
	secondaryDirWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), s.secondaryDirection)
	secondaryTan1, secondaryTan2 := secondaryDirWS.Normal()

	s.jacobian1 = physics.PairJacobian{
		Primary: physics.Jacobian{
			SlopeVelocity:        dprec.ZeroVec3(),
			SlopeAngularVelocity: dprec.Vec3Cross(primaryDirWS, secondaryTan1),
		},
		Secondary: physics.Jacobian{
			SlopeVelocity:        dprec.ZeroVec3(),
			SlopeAngularVelocity: dprec.Vec3Cross(secondaryTan1, primaryDirWS),
		},
	}
	s.jacobian2 = physics.PairJacobian{
		Primary: physics.Jacobian{
			SlopeVelocity:        dprec.ZeroVec3(),
			SlopeAngularVelocity: dprec.Vec3Cross(primaryDirWS, secondaryTan2),
		},
		Secondary: physics.Jacobian{
			SlopeVelocity:        dprec.ZeroVec3(),
			SlopeAngularVelocity: dprec.Vec3Cross(secondaryTan2, primaryDirWS),
		},
	}

	s.drift1 = dprec.Vec3Dot(primaryDirWS, secondaryTan1)
	s.drift2 = dprec.Vec3Dot(primaryDirWS, secondaryTan2)
}

func (s *MatchDirections) ApplyImpulses(ctx physics.DBSolverContext) {
	const beta = 0.2
	if dprec.Abs(s.drift1) > 0 {
		solution := ctx.JacobianImpulseSolution(s.jacobian1, s.drift1, 0.0)
		ctx.ApplyImpulseSolution(solution)
	}
	if dprec.Abs(s.drift2) > 0 {
		solution := ctx.JacobianImpulseSolution(s.jacobian2, s.drift2, 0.0)
		ctx.ApplyImpulseSolution(solution)
	}
}

func (s *MatchDirections) ApplyNudges(ctx physics.DBSolverContext) {}
