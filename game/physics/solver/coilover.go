package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*Coilover)(nil)

type Coilover struct {
	physics.NilDBConstraintSolver

	PrimaryAnchor dprec.Vec3
	FrequencyHz   float64
	DampingRatio  float64

	appliedLambda float64
}

func (c *Coilover) Reset(physics.DBSolverContext) {
	c.appliedLambda = 0.0
}

func (c *Coilover) CalculateImpulses(ctx physics.DBSolverContext) physics.DBImpulseSolution {
	primary := ctx.Primary
	secondary := ctx.Secondary
	firstRadiusWS := dprec.QuatVec3Rotation(primary.Orientation(), c.PrimaryAnchor)
	firstAnchorWS := dprec.Vec3Sum(primary.Position(), firstRadiusWS)
	secondAnchorWS := secondary.Position()
	deltaPosition := dprec.Vec3Diff(secondAnchorWS, firstAnchorWS)
	if deltaPosition.Length() < epsilon {
		return physics.DBImpulseSolution{}
	}
	drift := deltaPosition.Length()
	normal := dprec.BasisXVec3()
	if drift > epsilon {
		normal = dprec.UnitVec3(deltaPosition)
	}

	jacobian := physics.PairJacobian{
		Primary: physics.Jacobian{
			SlopeVelocity: dprec.NewVec3(
				-normal.X,
				-normal.Y,
				-normal.Z,
			),
			SlopeAngularVelocity: dprec.NewVec3(
				-(normal.Z*firstRadiusWS.Y - normal.Y*firstRadiusWS.Z),
				-(normal.X*firstRadiusWS.Z - normal.Z*firstRadiusWS.X),
				-(normal.Y*firstRadiusWS.X - normal.X*firstRadiusWS.Y),
			),
		},
		Secondary: physics.Jacobian{
			SlopeVelocity: dprec.NewVec3(
				normal.X,
				normal.Y,
				normal.Z,
			),
			SlopeAngularVelocity: dprec.ZeroVec3(),
		},
	}

	invertedEffectiveMass := jacobian.InverseEffectiveMass(primary, secondary)
	w := 2.0 * dprec.Pi * c.FrequencyHz
	dc := 2.0 * c.DampingRatio * w / invertedEffectiveMass
	k := w * w / invertedEffectiveMass

	gamma := 1.0 / (ctx.ElapsedSeconds * (dc + ctx.ElapsedSeconds*k))
	beta := ctx.ElapsedSeconds * k * gamma

	velocityLambda := jacobian.EffectiveVelocity(primary, secondary)
	lambda := -(velocityLambda + beta*drift + gamma*c.appliedLambda) / (invertedEffectiveMass + gamma)
	c.appliedLambda += lambda
	return jacobian.ImpulseSolution(lambda)
}
