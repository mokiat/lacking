package physics

import "github.com/mokiat/gomath/dprec"

// Jacobian represents the 1x6 Jacobian matrix of a single-body velocity
// constraint.
type Jacobian struct {
	SlopeVelocity        dprec.Vec3
	SlopeAngularVelocity dprec.Vec3
}

// EffectiveVelocity returns the amount of velocity of the body that is
// going in the wrong direction.
func (j Jacobian) EffectiveVelocity(body *Body) float64 {
	linear := dprec.Vec3Dot(j.SlopeVelocity, body.velocity)
	angular := dprec.Vec3Dot(j.SlopeAngularVelocity, body.angularVelocity)
	return linear + angular
}

// InverseEffectiveMass returns the inverse of the effective mass with which
// the body affects the constraint.
func (j Jacobian) InverseEffectiveMass(body *Body) float64 {
	linear := dprec.Vec3Dot(j.SlopeVelocity, j.SlopeVelocity) / body.mass
	angular := dprec.Vec3Dot(dprec.Mat3Vec3Prod(dprec.InverseMat3(body.momentOfInertia), j.SlopeAngularVelocity), j.SlopeAngularVelocity)
	return linear + angular
}

// ImpulseSolution returns an impulse solution based on the lambda impulse
// amount applied according to this Jacobian.
func (j Jacobian) ImpulseSolution(lambda float64) SBImpulseSolution {
	return SBImpulseSolution{
		Impulse:        dprec.Vec3Prod(j.SlopeVelocity, lambda),
		AngularImpulse: dprec.Vec3Prod(j.SlopeAngularVelocity, lambda),
	}
}

// NudgeSolution returns a nudge solution based on the lambda nudge amount
// applied according to this Jacobian.
func (j Jacobian) NudgeSolution(lambda float64) SBNudgeSolution {
	return SBNudgeSolution{
		Nudge:        dprec.Vec3Prod(j.SlopeVelocity, lambda),
		AngularNudge: dprec.Vec3Prod(j.SlopeAngularVelocity, lambda),
	}
}

// PairJacobian represents the 1x12 Jacobian matrix of a double-body velocity
// constraint.
type PairJacobian struct {
	Primary   Jacobian
	Secondary Jacobian
}

// EffectiveVelocity returns the amount of the combined velocities of the two
// bodies that is going in the wrong direction.
func (j PairJacobian) EffectiveVelocity(primary, secondary *Body) float64 {
	return j.Primary.EffectiveVelocity(primary) + j.Secondary.EffectiveVelocity(secondary)
}

// InverseEffectiveMass returns the inverse of the effective mass with which
// the two bodies affect the constraint.
func (j PairJacobian) InverseEffectiveMass(primary, secondary *Body) float64 {
	return j.Primary.InverseEffectiveMass(primary) + j.Secondary.InverseEffectiveMass(secondary)
}

// ImpulseSolution returns an impulse solution based on the lambda impulse
// amount applied according to this Jacobian.
func (j PairJacobian) ImpulseSolution(lambda float64) DBImpulseSolution {
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

// NudgeSolution returns a nudge solution based on the lambda nudge amount
// applied according to this Jacobian.
func (j PairJacobian) NudgeSolution(lambda float64) DBNudgeSolution {
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
