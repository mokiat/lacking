package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*Coilover)(nil)

type Coilover struct {
	physics.NilDBConstraintSolver

	PrimaryAnchor sprec.Vec3
	FrequencyHz   float32
	DampingRatio  float32

	appliedLambda float32
}

func (c *Coilover) Reset() {
	c.appliedLambda = 0.0
}

func (c *Coilover) CalculateImpulses(primary, secondary *physics.Body, ctx physics.ConstraintContext) physics.DBImpulseSolution {
	firstRadiusWS := sprec.QuatVec3Rotation(primary.Orientation(), c.PrimaryAnchor)
	firstAnchorWS := sprec.Vec3Sum(primary.Position(), firstRadiusWS)
	secondAnchorWS := secondary.Position()
	deltaPosition := sprec.Vec3Diff(secondAnchorWS, firstAnchorWS)
	if deltaPosition.Length() < epsilon {
		return physics.DBImpulseSolution{}
	}
	drift := deltaPosition.Length()
	normal := sprec.BasisXVec3()
	if drift > epsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}

	jacobian := physics.PairJacobian{
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
	}

	invertedEffectiveMass := jacobian.InverseEffectiveMass(primary, secondary)
	w := 2.0 * sprec.Pi * c.FrequencyHz
	dc := 2.0 * c.DampingRatio * w / invertedEffectiveMass
	k := w * w / invertedEffectiveMass

	gamma := 1.0 / (ctx.ElapsedSeconds * (dc + ctx.ElapsedSeconds*k))
	beta := ctx.ElapsedSeconds * k * gamma

	velocityLambda := jacobian.EffectiveVelocity(primary, secondary)
	lambda := -(velocityLambda + beta*drift + gamma*c.appliedLambda) / (invertedEffectiveMass + gamma)
	c.appliedLambda += lambda
	return jacobian.ImpulseSolution(primary, secondary, lambda)
}
