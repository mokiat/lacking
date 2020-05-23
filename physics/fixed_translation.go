package physics

import "github.com/mokiat/gomath/sprec"

type FixedTranslationConstraint struct {
	NilConstraint
	Fixture sprec.Vec3
	Body    *Body
}

func (c FixedTranslationConstraint) ApplyImpulse(ctx Context) {
	result := c.Calculate()
	if sprec.Abs(result.Drift) > epsilon {
		result.Jacobian.CorrectVelocity(c.Body)
	}
}

func (c FixedTranslationConstraint) ApplyNudge(ctx Context) {
	result := c.Calculate()
	if sprec.Abs(result.Drift) > epsilon {
		result.Jacobian.CorrectPosition(c.Body, result.Drift)
	}
}

func (c FixedTranslationConstraint) Calculate() FixedTranslationConstraintResult {
	deltaPosition := sprec.Vec3Diff(c.Body.Position, c.Fixture)
	normal := sprec.BasisXVec3()
	if deltaPosition.SqrLength() > sqrEpsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}

	return FixedTranslationConstraintResult{
		Jacobian: Jacobian{
			SlopeVelocity: sprec.NewVec3(
				normal.X,
				normal.Y,
				normal.Z,
			),
			SlopeAngularVelocity: sprec.ZeroVec3(),
		},
		Drift: deltaPosition.Length(),
	}
}

type FixedTranslationConstraintResult struct {
	Jacobian Jacobian
	Drift    float32
}
