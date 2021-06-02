package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*MatchRotation)(nil)

// NewMatchRotation creates a new MatchRotation constraint solver.
func NewMatchRotation() *MatchRotation {
	return &MatchRotation{
		xAxis: NewMatchAxis().
			SetPrimaryAxis(sprec.BasisXVec3()).
			SetSecondaryAxis(sprec.BasisXVec3()),
		yAxis: NewMatchAxis().
			SetPrimaryAxis(sprec.BasisYVec3()).
			SetSecondaryAxis(sprec.BasisYVec3()),
	}
}

// NewMatchRotation represents the solution for a constraint
// that keeps two bodies oriented in the same direction on
// all axis.
type MatchRotation struct {
	xAxis *MatchAxis
	yAxis *MatchAxis
}

func (r *MatchRotation) Reset(ctx physics.DBSolverContext) {
	r.xAxis.Reset(ctx)
	r.yAxis.Reset(ctx)
}

func (r *MatchRotation) CalculateImpulses(ctx physics.DBSolverContext) physics.DBImpulseSolution {
	xSolution := r.xAxis.CalculateImpulses(ctx)
	ySolution := r.yAxis.CalculateImpulses(ctx)
	return physics.DBImpulseSolution{
		Primary: physics.SBImpulseSolution{
			Impulse:        sprec.Vec3Sum(xSolution.Primary.Impulse, ySolution.Primary.Impulse),
			AngularImpulse: sprec.Vec3Sum(xSolution.Primary.AngularImpulse, ySolution.Primary.AngularImpulse),
		},
		Secondary: physics.SBImpulseSolution{
			Impulse:        sprec.Vec3Sum(xSolution.Secondary.Impulse, ySolution.Secondary.Impulse),
			AngularImpulse: sprec.Vec3Sum(xSolution.Secondary.AngularImpulse, ySolution.Secondary.AngularImpulse),
		},
	}
}

func (r *MatchRotation) CalculateNudges(ctx physics.DBSolverContext) physics.DBNudgeSolution {
	xSolution := r.xAxis.CalculateNudges(ctx)
	ySolution := r.yAxis.CalculateNudges(ctx)
	return physics.DBNudgeSolution{
		Primary: physics.SBNudgeSolution{
			Nudge:        sprec.Vec3Sum(xSolution.Primary.Nudge, ySolution.Primary.Nudge),
			AngularNudge: sprec.Vec3Sum(xSolution.Primary.AngularNudge, ySolution.Primary.AngularNudge),
		},
		Secondary: physics.SBNudgeSolution{
			Nudge:        sprec.Vec3Sum(xSolution.Secondary.Nudge, ySolution.Secondary.Nudge),
			AngularNudge: sprec.Vec3Sum(xSolution.Secondary.AngularNudge, ySolution.Secondary.AngularNudge),
		},
	}
}
