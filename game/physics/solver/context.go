package solver

// // PairContext contains information related to double-object constraint
// // processing.
// type PairContext struct {
// 	DeltaTime   float64
// 	ImpulseBeta float64
// 	NudgeBeta   float64

// 	Target *Placeholder
// 	Source *Placeholder
// }

// // JacobianImpulseLambda returns the impulse lambda for the specified
// // constraint Jacobian, positional drift and restitution.
// func (c PairContext) JacobianImpulseLambda(jacobian PairJacobian, drift, restitution float64) float64 {
// 	effMass := jacobian.InverseEffectiveMass(c.Target, c.Source)
// 	if effMass < Epsilon {
// 		return 0.0
// 	}
// 	effVelocity := jacobian.EffectiveVelocity(c.Target, c.Source)
// 	restitutionClamp := RestitutionClamp(effVelocity)
// 	baumgarte := c.ImpulseBeta * drift / c.DeltaTime
// 	return -((1+restitution*restitutionClamp)*effVelocity + baumgarte) / effMass
// }

// // JacobianNudgeLambda returns the nudge lambda for the specified
// // constraint Jacobian and positional drift.
// func (c PairContext) JacobianNudgeLambda(jacobian PairJacobian, drift float64) float64 {
// 	effMass := jacobian.InverseEffectiveMass(c.Target, c.Source)
// 	if effMass < Epsilon {
// 		return 0.0
// 	}
// 	return -c.NudgeBeta * drift / effMass
// }

// // JacobianImpulseSolution returns an impulse solution based on the specified
// // constraint Jacobian, positional drift and restitution.
// func (c PairContext) JacobianImpulseSolution(jacobian PairJacobian, drift, restitution float64) PairImpulse {
// 	lambda := c.JacobianImpulseLambda(jacobian, drift, restitution)
// 	return jacobian.Impulse(lambda)
// }

// // JacobianNudgeSolution returns a nudge solution based on the specified
// // constraint Jacobian and positional drift.
// func (c PairContext) JacobianNudgeSolution(jacobian PairJacobian, drift float64) PairNudge {
// 	lambda := c.JacobianNudgeLambda(jacobian, drift)
// 	return jacobian.Nudge(lambda)
// }
