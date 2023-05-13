package solver

// Context contains information related to single-object constraint
// processing.
type Context struct {
	DeltaTime   float64
	ImpulseBeta float64
	NudgeBeta   float64

	Target *Placeholder
}

// JacobianImpulseLambda returns the impulse lambda for the specified
// constraint Jacobian, positional drift and restitution.
func (c Context) JacobianImpulseLambda(jacobian Jacobian, drift, restitution float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Target)
	if effMass < Epsilon {
		return 0.0
	}
	effVelocity := jacobian.EffectiveVelocity(c.Target)
	resitutionClamp := RestitutionClamp(effVelocity)
	baumgarte := c.ImpulseBeta * drift / c.DeltaTime
	return -((1+restitution*resitutionClamp)*effVelocity + baumgarte) / effMass
}

// JacobianNudgeLambda returns the nudge lambda for the specified
// constraint Jacobian and positional drift.
func (c Context) JacobianNudgeLambda(jacobian Jacobian, drift float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Target)
	if effMass < Epsilon {
		return 0.0
	}
	return -c.NudgeBeta * drift / effMass
}

// JacobianImpulseSolution returns an impulse solution based on the specified
// constraint Jacobian, positional drift and restitution.
func (c Context) JacobianImpulseSolution(jacobian Jacobian, drift, restitution float64) Impulse {
	lambda := c.JacobianImpulseLambda(jacobian, drift, restitution)
	return jacobian.Impulse(lambda)
}

// JacobianNudgeSolution returns a nudge solution based on the specified
// constraint Jacobian and positional drift.
func (c Context) JacobianNudgeSolution(jacobian Jacobian, drift float64) Nudge {
	lambda := c.JacobianNudgeLambda(jacobian, drift)
	return jacobian.Nudge(lambda)
}

// PairContext contains information related to double-object constraint
// processing.
type PairContext struct {
	DeltaTime   float64
	ImpulseBeta float64
	NudgeBeta   float64

	Target *Placeholder
	Source *Placeholder
}

// JacobianImpulseLambda returns the impulse lambda for the specified
// constraint Jacobian, positional drift and restitution.
func (c PairContext) JacobianImpulseLambda(jacobian PairJacobian, drift, restitution float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Target, c.Source)
	if effMass < Epsilon {
		return 0.0
	}
	effVelocity := jacobian.EffectiveVelocity(c.Target, c.Source)
	resitutionClamp := RestitutionClamp(effVelocity)
	baumgarte := c.ImpulseBeta * drift / c.DeltaTime
	return -((1+restitution*resitutionClamp)*effVelocity + baumgarte) / effMass
}

// JacobianNudgeLambda returns the nudge lambda for the specified
// constraint Jacobian and positional drift.
func (c PairContext) JacobianNudgeLambda(jacobian PairJacobian, drift float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Target, c.Source)
	if effMass < Epsilon {
		return 0.0
	}
	return -c.NudgeBeta * drift / effMass
}

// JacobianImpulseSolution returns an impulse solution based on the specified
// constraint Jacobian, positional drift and restitution.
func (c PairContext) JacobianImpulseSolution(jacobian PairJacobian, drift, restitution float64) PairImpulse {
	lambda := c.JacobianImpulseLambda(jacobian, drift, restitution)
	return jacobian.Impulse(lambda)
}

// JacobianNudgeSolution returns a nudge solution based on the specified
// constraint Jacobian and positional drift.
func (c PairContext) JacobianNudgeSolution(jacobian PairJacobian, drift float64) PairNudge {
	lambda := c.JacobianNudgeLambda(jacobian, drift)
	return jacobian.Nudge(lambda)
}
