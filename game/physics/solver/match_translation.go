package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.ConstraintSolver = (*MatchTranslation)(nil)

type MatchTranslation struct {
	physics.NilConstraintSolver
	PrimaryAnchor sprec.Vec3
	IgnoreX       bool
	IgnoreY       bool
	IgnoreZ       bool
}

func (t *MatchTranslation) CalculateImpulses(primary, secondary *physics.Body, elapsedSeconds float32) physics.ConstraintImpulseSolution {
	jacobian, drift := t.calculate(primary, secondary)
	if sprec.Abs(drift) < epsilon {
		return physics.ConstraintImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(primary, secondary)
	return jacobian.ImpulseSolution(primary, secondary, lambda)
}

func (t *MatchTranslation) CalculateNudges(primary, secondary *physics.Body, elapsedSeconds float32) physics.ConstraintNudgeSolution {
	jacobian, drift := t.calculate(primary, secondary)
	if sprec.Abs(drift) < epsilon {
		return physics.ConstraintNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(primary, secondary, drift)
	return jacobian.NudgeSolution(primary, secondary, lambda)
}

func (t *MatchTranslation) calculate(primary, secondary *physics.Body) (physics.PairJacobian, float32) {
	firstRadiusWS := sprec.QuatVec3Rotation(primary.Orientation(), t.PrimaryAnchor)
	firstAnchorWS := sprec.Vec3Sum(primary.Position(), firstRadiusWS)
	deltaPosition := sprec.Vec3Diff(secondary.Position(), firstAnchorWS)
	if t.IgnoreX {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(primary.Orientation().OrientationX(), sprec.Vec3Dot(deltaPosition, primary.Orientation().OrientationX())))
	}
	if t.IgnoreY {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(primary.Orientation().OrientationY(), sprec.Vec3Dot(deltaPosition, primary.Orientation().OrientationY())))
	}
	if t.IgnoreZ {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(primary.Orientation().OrientationZ(), sprec.Vec3Dot(deltaPosition, primary.Orientation().OrientationZ())))
	}
	normal := sprec.BasisXVec3()
	if deltaPosition.SqrLength() > sqrEpsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}
	return physics.PairJacobian{
			Primary: physics.Jacobian{
				SlopeVelocity: sprec.NewVec3(
					-normal.X,
					-normal.Y,
					-normal.Z,
				),
				SlopeAngularVelocity: sprec.NewVec3(
					-(normal.Z*firstRadiusWS.Y - normal.Y*firstRadiusWS.Z),
					-(normal.X*firstRadiusWS.Z - normal.Z*firstRadiusWS.X),
					-(normal.Y*firstRadiusWS.X - normal.X*firstRadiusWS.Y),
				),
			},
			Secondary: physics.Jacobian{
				SlopeVelocity: sprec.NewVec3(
					normal.X,
					normal.Y,
					normal.Z,
				),
				SlopeAngularVelocity: sprec.ZeroVec3(),
			},
		},
		deltaPosition.Length()
}
