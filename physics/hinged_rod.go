package physics

import "github.com/mokiat/gomath/sprec"

type HingedRodConstraint struct {
	NilConstraint
	FirstBody        *Body
	FirstBodyAnchor  sprec.Vec3
	SecondBody       *Body
	SecondBodyAnchor sprec.Vec3
	Length           float32
}

func (c HingedRodConstraint) ApplyImpulse(ctx Context) {
	result := c.Calculate()
	if sprec.Abs(result.Drift) > epsilon {
		result.Jacobian.CorrectVelocity(c.FirstBody, c.SecondBody)
	}
}

func (c HingedRodConstraint) ApplyNudge(ctx Context) {
	result := c.Calculate()
	if sprec.Abs(result.Drift) > epsilon {
		result.Jacobian.CorrectPosition(c.FirstBody, c.SecondBody, result.Drift)
	}
}

func (c HingedRodConstraint) Calculate() HingedRodConstraintResult {
	firstRadiusWS := sprec.QuatVec3Rotation(c.FirstBody.Orientation, c.FirstBodyAnchor)
	secondRadiusWS := sprec.QuatVec3Rotation(c.SecondBody.Orientation, c.SecondBodyAnchor)
	firstAnchorWS := sprec.Vec3Sum(c.FirstBody.Position, firstRadiusWS)
	secondAnchorWS := sprec.Vec3Sum(c.SecondBody.Position, secondRadiusWS)
	deltaPosition := sprec.Vec3Diff(secondAnchorWS, firstAnchorWS)
	normal := sprec.BasisXVec3()
	if deltaPosition.SqrLength() > sqrEpsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}

	return HingedRodConstraintResult{
		Jacobian: PairJacobian{
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
				SlopeAngularVelocity: sprec.NewVec3(
					normal.Z*secondRadiusWS.Y-normal.Y*secondRadiusWS.Z,
					normal.X*secondRadiusWS.Z-normal.Z*secondRadiusWS.X,
					normal.Y*secondRadiusWS.X-normal.X*secondRadiusWS.Y,
				),
			},
		},
		Drift: deltaPosition.Length() - c.Length,
	}
}

type HingedRodConstraintResult struct {
	Jacobian PairJacobian
	Drift    float32
}
