package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

const (
	epsilon    = float32(0.001)
	sqrEpsilon = epsilon * epsilon
)

var _ physics.DBConstraintSolver = (*MatchAxis)(nil)

type MatchAxis struct {
	physics.NilDBConstraintSolver
	PrimaryAxis   sprec.Vec3
	SecondaryAxis sprec.Vec3
}

func (a *MatchAxis) CalculateImpulses(primary, secondary *physics.Body, ctx physics.ConstraintContext) physics.DBImpulseSolution {
	jacobian, drift := a.calculate(primary, secondary)
	if sprec.Abs(drift) < epsilon {
		return physics.DBImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(primary, secondary)
	return jacobian.ImpulseSolution(primary, secondary, lambda)
}

func (a *MatchAxis) CalculateNudges(primary, secondary *physics.Body, ctx physics.ConstraintContext) physics.DBNudgeSolution {
	jacobian, drift := a.calculate(primary, secondary)
	if sprec.Abs(drift) < epsilon {
		return physics.DBNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(primary, secondary, drift)
	return jacobian.NudgeSolution(primary, secondary, lambda)
}

func (a *MatchAxis) calculate(primary, secondary *physics.Body) (physics.PairJacobian, float32) {
	// FIXME: Does not handle when axis are pointing in opposite directions
	firstAxisWS := sprec.QuatVec3Rotation(primary.Orientation(), a.PrimaryAxis)
	secondAxisWS := sprec.QuatVec3Rotation(secondary.Orientation(), a.SecondaryAxis)
	cross := sprec.Vec3Cross(firstAxisWS, secondAxisWS)
	return physics.PairJacobian{
			Primary: physics.Jacobian{
				SlopeVelocity:        sprec.ZeroVec3(),
				SlopeAngularVelocity: sprec.InverseVec3(cross),
			},
			Secondary: physics.Jacobian{
				SlopeVelocity:        sprec.ZeroVec3(),
				SlopeAngularVelocity: cross,
			},
		},
		cross.Length()
}
