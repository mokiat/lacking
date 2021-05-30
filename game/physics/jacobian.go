package physics

import "github.com/mokiat/gomath/sprec"

const driftCorrectionAmount = float32(0.01) // TODO: Configurable?

type Jacobian struct {
	SlopeVelocity        sprec.Vec3
	SlopeAngularVelocity sprec.Vec3
}

func (j Jacobian) EffectiveVelocity(body *Body) float32 {
	return sprec.Vec3Dot(j.SlopeVelocity, body.velocity) +
		sprec.Vec3Dot(j.SlopeAngularVelocity, body.angularVelocity)
}

func (j Jacobian) InverseEffectiveMass(body *Body) float32 {
	return sprec.Vec3Dot(j.SlopeVelocity, j.SlopeVelocity)/body.mass +
		sprec.Vec3Dot(sprec.Mat3Vec3Prod(sprec.InverseMat3(body.momentOfInertia), j.SlopeAngularVelocity), j.SlopeAngularVelocity)
}

func (j Jacobian) ImpulseLambda(body *Body) float32 {
	return -j.EffectiveVelocity(body) / j.InverseEffectiveMass(body)
}

func (j Jacobian) ImpulseSolution(body *Body, lambda float32) ConstraintImpulseSolution {
	return ConstraintImpulseSolution{
		PrimaryImpulse:        sprec.Vec3Prod(j.SlopeVelocity, lambda),
		PrimaryAngularImpulse: sprec.Vec3Prod(j.SlopeAngularVelocity, lambda),
	}
}

func (j Jacobian) NudgeLambda(body *Body, drift float32) float32 {
	return -driftCorrectionAmount * drift / j.InverseEffectiveMass(body)
}

func (j Jacobian) NudgeSolution(body *Body, lambda float32) ConstraintNudgeSolution {
	return ConstraintNudgeSolution{
		PrimaryNudge:        sprec.Vec3Prod(j.SlopeVelocity, lambda),
		PrimaryAngularNudge: sprec.Vec3Prod(j.SlopeAngularVelocity, lambda),
	}
}

type PairJacobian struct {
	Primary   Jacobian
	Secondary Jacobian
}

func (j PairJacobian) ImpulseLambda(primary, secondary *Body) float32 {
	return -j.EffectiveVelocity(primary, secondary) / j.InverseEffectiveMass(primary, secondary)
}

func (j PairJacobian) ImpulseSolution(primary, secondary *Body, lambda float32) ConstraintImpulseSolution {
	return ConstraintImpulseSolution{
		PrimaryImpulse:          sprec.Vec3Prod(j.Primary.SlopeVelocity, lambda),
		PrimaryAngularImpulse:   sprec.Vec3Prod(j.Primary.SlopeAngularVelocity, lambda),
		SecondaryImpulse:        sprec.Vec3Prod(j.Secondary.SlopeVelocity, lambda),
		SecondaryAngularImpulse: sprec.Vec3Prod(j.Secondary.SlopeAngularVelocity, lambda),
	}
}

func (j PairJacobian) NudgeLambda(primary, secondary *Body, drift float32) float32 {
	return -driftCorrectionAmount * drift / j.InverseEffectiveMass(primary, secondary)
}

func (j PairJacobian) NudgeSolution(primary, secondary *Body, lambda float32) ConstraintNudgeSolution {
	return ConstraintNudgeSolution{
		PrimaryNudge:          sprec.Vec3Prod(j.Primary.SlopeVelocity, lambda),
		PrimaryAngularNudge:   sprec.Vec3Prod(j.Primary.SlopeAngularVelocity, lambda),
		SecondaryNudge:        sprec.Vec3Prod(j.Secondary.SlopeVelocity, lambda),
		SecondaryAngularNudge: sprec.Vec3Prod(j.Secondary.SlopeAngularVelocity, lambda),
	}
}

func (j PairJacobian) EffectiveVelocity(primary, secondary *Body) float32 {
	return j.Primary.EffectiveVelocity(primary) + j.Secondary.EffectiveVelocity(secondary)
}

func (j PairJacobian) InverseEffectiveMass(primary, secondary *Body) float32 {
	return j.Primary.InverseEffectiveMass(primary) + j.Secondary.InverseEffectiveMass(secondary)
}
