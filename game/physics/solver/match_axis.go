package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

const (
	epsilon    = float32(0.001)
	sqrEpsilon = epsilon * epsilon
)

var _ physics.ConstraintSolver = (*MatchAxis)(nil)

type MatchAxis struct {
	physics.NilConstraintSolver
	PrimaryAxis   sprec.Vec3
	SecondaryAxis sprec.Vec3
}

func (a *MatchAxis) CalculateImpulses(primary, secondary *physics.Body, elapsedSeconds float32) physics.ConstraintImpulseSolution {
	jacobian, drift := a.calculate(primary, secondary)
	if sprec.Abs(drift) < epsilon {
		return physics.ConstraintImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(primary, secondary)
	return jacobian.ImpulseSolution(primary, secondary, lambda)
}

func (a *MatchAxis) CalculateNudges(primary, secondary *physics.Body, elapsedSeconds float32) physics.ConstraintNudgeSolution {
	jacobian, drift := a.calculate(primary, secondary)
	if sprec.Abs(drift) < epsilon {
		return physics.ConstraintNudgeSolution{}
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
