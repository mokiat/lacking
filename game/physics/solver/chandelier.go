package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.ConstraintSolver = (*Chandelier)(nil)

type Chandelier struct {
	physics.NilConstraintSolver

	Fixture    sprec.Vec3
	BodyAnchor sprec.Vec3
	Length     float32
}

func (c *Chandelier) CalculateImpulses(primary, secondary *physics.Body, elapsedSeconds float32) physics.ConstraintImpulseSolution {
	jacobian, drift := c.calculate(primary)
	if sprec.Abs(drift) < epsilon {
		return physics.ConstraintImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(primary)
	return jacobian.ImpulseSolution(primary, lambda)
}

func (c *Chandelier) CalculateNudges(primary, secondary *physics.Body, elapsedSeconds float32) physics.ConstraintNudgeSolution {
	jacobian, drift := c.calculate(primary)
	if sprec.Abs(drift) < epsilon {
		return physics.ConstraintNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(primary, drift)
	return jacobian.NudgeSolution(primary, lambda)
}

func (c *Chandelier) calculate(body *physics.Body) (physics.Jacobian, float32) {
	anchorWS := sprec.Vec3Sum(body.Position(), sprec.QuatVec3Rotation(body.Orientation(), c.BodyAnchor))
	radiusWS := sprec.Vec3Diff(anchorWS, body.Position())
	deltaPosition := sprec.Vec3Diff(anchorWS, c.Fixture)
	normal := sprec.BasisXVec3()
	if deltaPosition.SqrLength() > sqrEpsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}
	return physics.Jacobian{
			SlopeVelocity: sprec.NewVec3(
				normal.X,
				normal.Y,
				normal.Z,
			),
			SlopeAngularVelocity: sprec.NewVec3(
				normal.Z*radiusWS.Y-normal.Y*radiusWS.Z,
				normal.X*radiusWS.Z-normal.Z*radiusWS.X,
				normal.Y*radiusWS.X-normal.X*radiusWS.Y,
			),
		},
		deltaPosition.Length() - c.Length
}
