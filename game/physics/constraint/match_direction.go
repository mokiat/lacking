package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewMatchDirections creates a new MatchDirections constraint solver.
func NewMatchDirections() *MatchDirections {
	return &MatchDirections{
		primaryDirection:   dprec.BasisYVec3(),
		secondaryDirection: dprec.BasisYVec3(),
	}
}

var _ solver.PairConstraint = (*MatchDirections)(nil)

// MatchDirections represents the solution for a constraint
// that keeps the direction of two bodies pointing in the same
// direction.
type MatchDirections struct {
	primaryDirection   dprec.Vec3
	secondaryDirection dprec.Vec3

	jacobian1 solver.PairJacobian
	jacobian2 solver.PairJacobian
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

func (s *MatchDirections) Reset(ctx solver.PairContext) {
	primaryDirWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.primaryDirection)
	secondaryDirWS := dprec.QuatVec3Rotation(ctx.Source.Rotation(), s.secondaryDirection)
	secondaryNorm1 := dprec.NormalVec3(secondaryDirWS)
	secondaryNorm2 := dprec.Vec3Cross(secondaryDirWS, secondaryNorm1)

	// FIXME: This jacobian converges better than the original one-tier
	// but produces a wrong result if the second object flips all the way
	// around.
	s.jacobian1 = solver.PairJacobian{
		Target: solver.Jacobian{
			LinearSlope:  dprec.ZeroVec3(),
			AngularSlope: dprec.Vec3Cross(primaryDirWS, secondaryNorm1),
		},
		Source: solver.Jacobian{
			LinearSlope:  dprec.ZeroVec3(),
			AngularSlope: dprec.Vec3Cross(secondaryNorm1, primaryDirWS),
		},
	}
	s.jacobian2 = solver.PairJacobian{
		Target: solver.Jacobian{
			LinearSlope:  dprec.ZeroVec3(),
			AngularSlope: dprec.Vec3Cross(primaryDirWS, secondaryNorm2),
		},
		Source: solver.Jacobian{
			LinearSlope:  dprec.ZeroVec3(),
			AngularSlope: dprec.Vec3Cross(secondaryNorm2, primaryDirWS),
		},
	}

	s.drift1 = dprec.Vec3Dot(primaryDirWS, secondaryNorm1)
	s.drift2 = dprec.Vec3Dot(primaryDirWS, secondaryNorm2)
}

func (s *MatchDirections) ApplyImpulses(ctx solver.PairContext) {
	solution := ctx.JacobianImpulseSolution(s.jacobian1, s.drift1, 0.0)
	ctx.Target.ApplyImpulse(solution.Target)
	ctx.Source.ApplyImpulse(solution.Source)
	solution = ctx.JacobianImpulseSolution(s.jacobian2, s.drift2, 0.0)
	ctx.Target.ApplyImpulse(solution.Target)
	ctx.Source.ApplyImpulse(solution.Source)
}

func (s *MatchDirections) ApplyNudges(ctx solver.PairContext) {
	solution := ctx.JacobianNudgeSolution(s.jacobian1, s.drift1)
	ctx.Target.ApplyNudge(solution.Target)
	ctx.Source.ApplyNudge(solution.Source)
	solution = ctx.JacobianNudgeSolution(s.jacobian2, s.drift2)
	ctx.Target.ApplyNudge(solution.Target)
	ctx.Source.ApplyNudge(solution.Source)
}
