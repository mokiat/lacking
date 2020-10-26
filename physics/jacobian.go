package physics

import "github.com/mokiat/gomath/sprec"

const driftCorrectionAmount = float32(0.01) // TODO: Configurable?

type Jacobian struct {
	SlopeVelocity        sprec.Vec3
	SlopeAngularVelocity sprec.Vec3
}

func (j Jacobian) EffectiveVelocity(body *Body) float32 {
	return sprec.Vec3Dot(j.SlopeVelocity, body.Velocity) +
		sprec.Vec3Dot(j.SlopeAngularVelocity, body.AngularVelocity)
}

func (j Jacobian) InverseEffectiveMass(body *Body) float32 {
	return sprec.Vec3Dot(j.SlopeVelocity, j.SlopeVelocity)/body.Mass +
		sprec.Vec3Dot(sprec.Mat3Vec3Prod(sprec.InverseMat3(body.MomentOfInertia), j.SlopeAngularVelocity), j.SlopeAngularVelocity)
}

func (j Jacobian) ApplyImpulse(body *Body, lambda float32) {
	body.ApplyImpulse(sprec.Vec3Prod(j.SlopeVelocity, lambda))
	body.ApplyAngularImpulse(sprec.Vec3Prod(j.SlopeAngularVelocity, lambda))
}

func (j Jacobian) ApplyNudge(body *Body, lambda float32) {
	body.ApplyNudge(sprec.Vec3Prod(j.SlopeVelocity, lambda))
	body.ApplyAngularNudge(sprec.Vec3Prod(j.SlopeAngularVelocity, lambda))
}

func (j Jacobian) CorrectVelocity(body *Body) {
	lambda := -j.EffectiveVelocity(body) / j.InverseEffectiveMass(body)
	j.ApplyImpulse(body, lambda)
}

func (j Jacobian) CorrectPosition(body *Body, drift float32) {
	lambda := -driftCorrectionAmount * drift / j.InverseEffectiveMass(body)
	j.ApplyNudge(body, lambda)
}

type PairJacobian struct {
	First  Jacobian
	Second Jacobian
}

func (j PairJacobian) EffectiveVelocity(firstBody, secondBody *Body) float32 {
	return j.First.EffectiveVelocity(firstBody) + j.Second.EffectiveVelocity(secondBody)
}

func (j PairJacobian) InverseEffectiveMass(firstBody, secondBody *Body) float32 {
	return j.First.InverseEffectiveMass(firstBody) + j.Second.InverseEffectiveMass(secondBody)
}

func (j PairJacobian) ApplyImpulse(firstBody, secondBody *Body, lambda float32) {
	firstBody.ApplyImpulse(sprec.Vec3Prod(j.First.SlopeVelocity, lambda))
	firstBody.ApplyAngularImpulse(sprec.Vec3Prod(j.First.SlopeAngularVelocity, lambda))
	secondBody.ApplyImpulse(sprec.Vec3Prod(j.Second.SlopeVelocity, lambda))
	secondBody.ApplyAngularImpulse(sprec.Vec3Prod(j.Second.SlopeAngularVelocity, lambda))
}

func (j PairJacobian) ApplyNudge(firstBody, secondBody *Body, lambda float32) {
	firstBody.ApplyNudge(sprec.Vec3Prod(j.First.SlopeVelocity, lambda))
	firstBody.ApplyAngularNudge(sprec.Vec3Prod(j.First.SlopeAngularVelocity, lambda))
	secondBody.ApplyNudge(sprec.Vec3Prod(j.Second.SlopeVelocity, lambda))
	secondBody.ApplyAngularNudge(sprec.Vec3Prod(j.Second.SlopeAngularVelocity, lambda))
}

func (j PairJacobian) CorrectVelocity(firstBody, secondBody *Body) {
	lambda := -j.EffectiveVelocity(firstBody, secondBody) / j.InverseEffectiveMass(firstBody, secondBody)
	j.ApplyImpulse(firstBody, secondBody, lambda)
}

func (j PairJacobian) CorrectPosition(firstBody, secondBody *Body, drift float32) {
	lambda := -driftCorrectionAmount * drift / j.InverseEffectiveMass(firstBody, secondBody)
	j.ApplyNudge(firstBody, secondBody, lambda)
}
