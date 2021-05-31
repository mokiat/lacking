package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.SBConstraintSolver = (*Chandelier)(nil)

type Chandelier struct {
	physics.NilSBConstraintSolver

	Fixture    sprec.Vec3
	BodyAnchor sprec.Vec3
	Length     float32
}

func (c *Chandelier) CalculateImpulses(body *physics.Body, ctx physics.ConstraintContext) physics.SBImpulseSolution {
	jacobian, drift := c.calculate(body)
	if sprec.Abs(drift) < epsilon {
		return physics.SBImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(body)
	return jacobian.ImpulseSolution(body, lambda)
}

func (c *Chandelier) CalculateNudges(body *physics.Body, ctx physics.ConstraintContext) physics.SBNudgeSolution {
	jacobian, drift := c.calculate(body)
	if sprec.Abs(drift) < epsilon {
		return physics.SBNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(body, drift)
	return jacobian.NudgeSolution(body, lambda)
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
