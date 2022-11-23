package physics

import "github.com/mokiat/gomath/dprec"

// TODO: Come up with some better value here. Try to get it to work correctly
// for values around 0.5 to 0.75.
const driftCorrectionAmount = float64(0.01)

type Jacobian struct {
	SlopeVelocity        dprec.Vec3
	SlopeAngularVelocity dprec.Vec3
}

func (j Jacobian) EffectiveVelocity(body *Body) float64 {
	return dprec.Vec3Dot(j.SlopeVelocity, body.velocity) +
		dprec.Vec3Dot(j.SlopeAngularVelocity, body.angularVelocity)
}

func (j Jacobian) InverseEffectiveMass(body *Body) float64 {
	return dprec.Vec3Dot(j.SlopeVelocity, j.SlopeVelocity)/body.mass +
		dprec.Vec3Dot(dprec.Mat3Vec3Prod(dprec.InverseMat3(body.momentOfInertia), j.SlopeAngularVelocity), j.SlopeAngularVelocity)
}

func (j Jacobian) ImpulseLambda(body *Body) float64 {
	// TODO: If effective velocity is zero, then don't calculate the inverse
	// effective mass since this collision is already solved (also
	// the inverse effective mass would likely be zero for some constraints).
	return -j.EffectiveVelocity(body) / j.InverseEffectiveMass(body)
}

func (j Jacobian) ImpulseSolution(body *Body, lambda float64) SBImpulseSolution {
	return SBImpulseSolution{
		Impulse:        dprec.Vec3Prod(j.SlopeVelocity, lambda),
		AngularImpulse: dprec.Vec3Prod(j.SlopeAngularVelocity, lambda),
	}
}

func (j Jacobian) NudgeLambda(body *Body, drift float64) float64 {
	// TODO: It drift is close to zero, then don't calculate the inverse
	// effective mass, since this constraint is solved (also
	// the inverse effective mass would likely be zero for some constraints).
	return -driftCorrectionAmount * drift / j.InverseEffectiveMass(body)
}

func (j Jacobian) NudgeSolution(body *Body, lambda float64) SBNudgeSolution {
	return SBNudgeSolution{
		Nudge:        dprec.Vec3Prod(j.SlopeVelocity, lambda),
		AngularNudge: dprec.Vec3Prod(j.SlopeAngularVelocity, lambda),
	}
}

type PairJacobian struct {
	Primary   Jacobian
	Secondary Jacobian
}

func (j PairJacobian) ImpulseLambda(primary, secondary *Body) float64 {
	// TODO: If effective velocity is zero, then don't calculate the inverse
	// effective mass since this collision is already solved (also
	// the inverse effective mass would likely be zero for some constraints).
	return -j.EffectiveVelocity(primary, secondary) / j.InverseEffectiveMass(primary, secondary)
}

func (j PairJacobian) ImpulseSolution(primary, secondary *Body, lambda float64) DBImpulseSolution {
	return DBImpulseSolution{
		Primary: SBImpulseSolution{
			Impulse:        dprec.Vec3Prod(j.Primary.SlopeVelocity, lambda),
			AngularImpulse: dprec.Vec3Prod(j.Primary.SlopeAngularVelocity, lambda),
		},
		Secondary: SBImpulseSolution{
			Impulse:        dprec.Vec3Prod(j.Secondary.SlopeVelocity, lambda),
			AngularImpulse: dprec.Vec3Prod(j.Secondary.SlopeAngularVelocity, lambda),
		},
	}
}

func (j PairJacobian) NudgeLambda(primary, secondary *Body, drift float64) float64 {
	// TODO: It drift is close to zero, then don't calculate the inverse
	// effective mass, since this constraint is solved (also
	// the inverse effective mass would likely be zero for some constraints).
	return -driftCorrectionAmount * drift / j.InverseEffectiveMass(primary, secondary)
}

func (j PairJacobian) NudgeSolution(primary, secondary *Body, lambda float64) DBNudgeSolution {
	return DBNudgeSolution{
		Primary: SBNudgeSolution{
			Nudge:        dprec.Vec3Prod(j.Primary.SlopeVelocity, lambda),
			AngularNudge: dprec.Vec3Prod(j.Primary.SlopeAngularVelocity, lambda),
		},
		Secondary: SBNudgeSolution{
			Nudge:        dprec.Vec3Prod(j.Secondary.SlopeVelocity, lambda),
			AngularNudge: dprec.Vec3Prod(j.Secondary.SlopeAngularVelocity, lambda),
		},
	}
}

func (j PairJacobian) EffectiveVelocity(primary, secondary *Body) float64 {
	return j.Primary.EffectiveVelocity(primary) + j.Secondary.EffectiveVelocity(secondary)
}

func (j PairJacobian) InverseEffectiveMass(primary, secondary *Body) float64 {
	return j.Primary.InverseEffectiveMass(primary) + j.Secondary.InverseEffectiveMass(secondary)
}
