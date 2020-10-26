package physics

import "github.com/mokiat/gomath/sprec"

type MatchAxisConstraint struct {
	NilConstraint
	FirstBody      *Body
	FirstBodyAxis  sprec.Vec3
	SecondBody     *Body
	SecondBodyAxis sprec.Vec3
}

func (c MatchAxisConstraint) ApplyImpulse(ctx Context) {
	result := c.Calculate()
	if sprec.Abs(result.Drift) > epsilon {
		result.Jacobian.CorrectVelocity(c.FirstBody, c.SecondBody)
	}
}

func (c MatchAxisConstraint) ApplyNudge(ctx Context) {
	result := c.Calculate()
	if sprec.Abs(result.Drift) > epsilon {
		result.Jacobian.CorrectPosition(c.FirstBody, c.SecondBody, result.Drift)
	}
}

func (c MatchAxisConstraint) Calculate() MatchAxisConstraintResult {
	// FIXME: Does not handle when axis are pointing in opposite directions
	firstAxisWS := sprec.QuatVec3Rotation(c.FirstBody.Orientation, c.FirstBodyAxis)
	secondAxisWS := sprec.QuatVec3Rotation(c.SecondBody.Orientation, c.SecondBodyAxis)
	cross := sprec.Vec3Cross(firstAxisWS, secondAxisWS)
	return MatchAxisConstraintResult{
		Jacobian: PairJacobian{
			First: Jacobian{
				SlopeVelocity:        sprec.ZeroVec3(),
				SlopeAngularVelocity: sprec.InverseVec3(cross),
			},
			Second: Jacobian{
				SlopeVelocity:        sprec.ZeroVec3(),
				SlopeAngularVelocity: cross,
			},
		},
		Drift: cross.Length(),
	}
}

type MatchAxisConstraintResult struct {
	Jacobian PairJacobian
	Drift    float32
}
