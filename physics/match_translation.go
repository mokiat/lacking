package physics

import "github.com/mokiat/gomath/sprec"

type MatchTranslationConstraint struct {
	NilConstraint
	FirstBody       *Body
	FirstBodyAnchor sprec.Vec3
	SecondBody      *Body
	IgnoreX         bool
	IgnoreY         bool
	IgnoreZ         bool
}

func (c MatchTranslationConstraint) ApplyImpulse(ctx Context) {
	result := c.Calculate()
	if sprec.Abs(result.Drift) > epsilon {
		result.Jacobian.CorrectVelocity(c.FirstBody, c.SecondBody)
	}
}

func (c MatchTranslationConstraint) ApplyNudge(ctx Context) {
	result := c.Calculate()
	if sprec.Abs(result.Drift) > epsilon {
		result.Jacobian.CorrectPosition(c.FirstBody, c.SecondBody, result.Drift)
	}
}

func (c MatchTranslationConstraint) Calculate() MatchTranslationResultConstraint {
	firstRadiusWS := sprec.QuatVec3Rotation(c.FirstBody.Orientation, c.FirstBodyAnchor)
	firstAnchorWS := sprec.Vec3Sum(c.FirstBody.Position, firstRadiusWS)
	deltaPosition := sprec.Vec3Diff(c.SecondBody.Position, firstAnchorWS)
	if c.IgnoreX {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(c.FirstBody.Orientation.OrientationX(), sprec.Vec3Dot(deltaPosition, c.FirstBody.Orientation.OrientationX())))
	}
	if c.IgnoreY {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(c.FirstBody.Orientation.OrientationY(), sprec.Vec3Dot(deltaPosition, c.FirstBody.Orientation.OrientationY())))
	}
	if c.IgnoreZ {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(c.FirstBody.Orientation.OrientationZ(), sprec.Vec3Dot(deltaPosition, c.FirstBody.Orientation.OrientationZ())))
	}
	normal := sprec.BasisXVec3()
	if deltaPosition.SqrLength() > sqrEpsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}

	return MatchTranslationResultConstraint{
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
				SlopeAngularVelocity: sprec.ZeroVec3(),
			},
		},
		Drift: deltaPosition.Length(),
	}
}

type MatchTranslationResultConstraint struct {
	Jacobian PairJacobian
	Drift    float32
}
