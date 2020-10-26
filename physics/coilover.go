package physics

import "github.com/mokiat/gomath/sprec"

type CoiloverConstraint struct {
	NilConstraint

	FirstBody       *Body
	FirstBodyAnchor sprec.Vec3
	SecondBody      *Body
	FrequencyHz     float32
	DampingRatio    float32

	appliedLambda float32
}

func (c *CoiloverConstraint) Reset() {
	c.appliedLambda = 0.0
}

func (c *CoiloverConstraint) ApplyImpulse(ctx Context) {
	firstRadiusWS := sprec.QuatVec3Rotation(c.FirstBody.Orientation, c.FirstBodyAnchor)
	firstAnchorWS := sprec.Vec3Sum(c.FirstBody.Position, firstRadiusWS)
	secondAnchorWS := c.SecondBody.Position
	deltaPosition := sprec.Vec3Diff(secondAnchorWS, firstAnchorWS)
	if deltaPosition.Length() < epsilon {
		return
	}
	drift := deltaPosition.Length()
	normal := sprec.BasisXVec3()
	if drift > epsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}

	jacobian := PairJacobian{
		First: Jacobian{
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
		Second: Jacobian{
			SlopeVelocity: sprec.NewVec3(
				normal.X,
				normal.Y,
				normal.Z,
			),
			SlopeAngularVelocity: sprec.ZeroVec3(),
		},
	}

	invertedEffectiveMass := jacobian.InverseEffectiveMass(c.FirstBody, c.SecondBody)
	w := 2.0 * sprec.Pi * c.FrequencyHz
	dc := 2.0 * c.DampingRatio * w / invertedEffectiveMass
	k := w * w / invertedEffectiveMass

	gamma := 1.0 / (ctx.ElapsedSeconds * (dc + ctx.ElapsedSeconds*k))
	beta := ctx.ElapsedSeconds * k * gamma

	velocityLambda := jacobian.EffectiveVelocity(c.FirstBody, c.SecondBody)
	lambda := -(velocityLambda + beta*drift + gamma*c.appliedLambda) / (invertedEffectiveMass + gamma)
	c.appliedLambda += lambda
	jacobian.ApplyImpulse(c.FirstBody, c.SecondBody, lambda)
}
