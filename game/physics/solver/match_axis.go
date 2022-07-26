package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*MatchAxis)(nil)

// NewMatchAxis creates a new MatchAxis constraint solver.
func NewMatchAxis() *MatchAxis {
	result := &MatchAxis{
		primaryAxis:   dprec.BasisXVec3(),
		secondaryAxis: dprec.BasisXVec3(),
	}
	result.DBJacobianConstraintSolver = physics.NewDBJacobianConstraintSolver(result.calculate)
	return result
}

// MatchAxis represents the solution for a constraint
// that keeps the axis of two bodies pointing in the same
// direction.
type MatchAxis struct {
	*physics.DBJacobianConstraintSolver

	primaryAxis   dprec.Vec3
	secondaryAxis dprec.Vec3
}

// PrimaryAxis returns the axis of the primary body that will be
// used in the alignment.
func (a *MatchAxis) PrimaryAxis() dprec.Vec3 {
	return a.primaryAxis
}

// SetPrimaryAxis changes the axis of the primary body to be used
// in alignments.
func (a *MatchAxis) SetPrimaryAxis(axis dprec.Vec3) *MatchAxis {
	a.primaryAxis = axis
	return a
}

// SecondaryAxis returns the axis of the secondary body that will be
// used in the alignment.
func (a *MatchAxis) SecondaryAxis() dprec.Vec3 {
	return a.secondaryAxis
}

// SetSecondaryAxis changes the axis of the secondary body to be
// used in alignments.
func (a *MatchAxis) SetSecondaryAxis(axis dprec.Vec3) *MatchAxis {
	a.secondaryAxis = axis
	return a
}

func (a *MatchAxis) calculate(ctx physics.DBSolverContext) (physics.PairJacobian, float64) {
	firstAxisWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), a.primaryAxis)
	secondAxisWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), a.secondaryAxis)
	cross := dprec.Vec3Cross(firstAxisWS, secondAxisWS)
	return physics.PairJacobian{
			Primary: physics.Jacobian{
				SlopeVelocity:        dprec.ZeroVec3(),
				SlopeAngularVelocity: dprec.InverseVec3(cross),
			},
			Secondary: physics.Jacobian{
				SlopeVelocity:        dprec.ZeroVec3(),
				SlopeAngularVelocity: cross,
			},
		},
		cross.Length()
}
