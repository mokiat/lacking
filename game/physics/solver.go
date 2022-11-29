package physics

import "github.com/mokiat/gomath/dprec"

var (
	// ImpulseIterationCount controls the number of iterations for impulse
	// solutions by the solvers.
	ImpulseIterationCount = 8
	// NudgeIterationCount controls the number of iterations for nudge
	// solutions by the solvers.
	NudgeIterationCount = 8

	// ImpulseDriftAdjustmentRatio controls the amount by which impulses should
	// try to correct positional drift.
	//
	// This is the `beta` coefficient in the Baumgarte stabilization approach.
	ImpulseDriftAdjustmentRatio = 0.2
	// NudgeDriftAdjustmentRatio controls the amount by which nudges should
	// try to correct positional drift.
	//
	// The value here is accumulated over all iterations. In fact, the total
	// remaining error is proportional to (1.0 - ratio) ^ iterations.
	//
	// Some error should be left in order to avoid jitters due to imprecise
	// integration of the correction and to leave some drift for the
	// impulse solution.
	NudgeDriftAdjustmentRatio = 0.2
)

// RestitutionClamp specifies the amount by which the restitution coefficient
// should be multiplied depending on the effective velocity.
//
// This clamp softens bounces when velocity is small.
func RestitutionClamp(effectiveVelocity float64) float64 {
	absEffectiveVelocity := dprec.Abs(effectiveVelocity)
	switch {
	case absEffectiveVelocity < 2.0:
		return 0.1
	case absEffectiveVelocity < 1.0:
		return 0.05
	case absEffectiveVelocity < 0.5:
		return 0.0
	default:
		return 1.0
	}
}

// SBImpulseSolution is a solution to a single-body constraint that
// contains the impulses that need to be applied to the body.
type SBImpulseSolution struct {
	Impulse        dprec.Vec3
	AngularImpulse dprec.Vec3
}

// SBNudgeSolution is a solution to a single-body constraint that
// contains the nudges that need to be applied to the body.
type SBNudgeSolution struct {
	Nudge        dprec.Vec3
	AngularNudge dprec.Vec3
}

// DBImpulseSolution is a solution to a constraint that indicates the impulses
// that need to be applied to the primary and secondary bodies.
type DBImpulseSolution struct {
	Primary   SBImpulseSolution
	Secondary SBImpulseSolution
}

// DBNudgeSolution is a solution to a constraint that indicates the nudges that
// need to be applied to the primary and secondary bodies.
type DBNudgeSolution struct {
	Primary   SBNudgeSolution
	Secondary SBNudgeSolution
}

// SBSolverContext contains information related to single-body constraint
// processing.
type SBSolverContext struct {
	Body           *Body
	ElapsedSeconds float64
}

// RestitutionCoefficient returns the restitution coefficient of the system.
func (c SBSolverContext) RestitutionCoefficient() float64 {
	return c.Body.restitutionCoefficient
}

// ApplyImpulseSolution applies the specified impulse solution to the body.
func (c SBSolverContext) ApplyImpulseSolution(solution SBImpulseSolution) {
	c.Body.applyImpulse(solution.Impulse)
	c.Body.applyAngularImpulse(solution.AngularImpulse)
}

// ApplyNudgeSolution applies the specified nudge solution to the body.
func (c SBSolverContext) ApplyNudgeSolution(solution SBNudgeSolution) {
	c.Body.applyNudge(solution.Nudge)
	c.Body.applyAngularNudge(solution.AngularNudge)
}

// JacobianImpulseLambda returns the impulse lambda for the specified
// constraint Jacobian, positional drift and restitution.
func (c SBSolverContext) JacobianImpulseLambda(jacobian Jacobian, drift, restitution float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Body)
	if effMass < epsilon {
		return 0.0
	}
	effVelocity := jacobian.EffectiveVelocity(c.Body)
	resitutionClamp := RestitutionClamp(effVelocity)
	baumgarte := ImpulseDriftAdjustmentRatio * drift / c.ElapsedSeconds
	return -((1+restitution*resitutionClamp)*effVelocity + baumgarte) / effMass
}

// JacobianNudgeLambda returns the nudge lambda for the specified
// constraint Jacobian and positional drift.
func (c SBSolverContext) JacobianNudgeLambda(jacobian Jacobian, drift float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Body)
	if effMass < epsilon {
		return 0.0
	}
	return -NudgeDriftAdjustmentRatio * drift / effMass
}

// JacobianImpulseSolution returns an impulse solution based on the specified
// constraint Jacobian, positional drift and restitution.
func (c SBSolverContext) JacobianImpulseSolution(jacobian Jacobian, drift, restitution float64) SBImpulseSolution {
	lambda := c.JacobianImpulseLambda(jacobian, drift, restitution)
	return jacobian.ImpulseSolution(lambda)
}

// JacobianNudgeSolution returns a nudge solution based on the specified
// constraint Jacobian and positional drift.
func (c SBSolverContext) JacobianNudgeSolution(jacobian Jacobian, drift float64) SBNudgeSolution {
	lambda := c.JacobianNudgeLambda(jacobian, drift)
	return jacobian.NudgeSolution(lambda)
}

// DBSolverContext contains information related to double-body constraint
// processing.
type DBSolverContext struct {
	Primary        *Body
	Secondary      *Body
	ElapsedSeconds float64
}

// RestitutionCoefficient returns the restitution coefficient of the system.
func (c DBSolverContext) RestitutionCoefficient() float64 {
	return c.Primary.restitutionCoefficient * c.Secondary.restitutionCoefficient
}

// ApplyImpulseSolution applies the specified impulse solution to the bodies.
func (c DBSolverContext) ApplyImpulseSolution(solution DBImpulseSolution) {
	c.Primary.applyImpulse(solution.Primary.Impulse)
	c.Primary.applyAngularImpulse(solution.Primary.AngularImpulse)
	c.Secondary.applyImpulse(solution.Secondary.Impulse)
	c.Secondary.applyAngularImpulse(solution.Secondary.AngularImpulse)
}

// ApplyNudgeSolution applies the specified nudge solution to the bodies.
func (c DBSolverContext) ApplyNudgeSolution(solution DBNudgeSolution) {
	c.Primary.applyNudge(solution.Primary.Nudge)
	c.Primary.applyAngularNudge(solution.Primary.AngularNudge)
	c.Secondary.applyNudge(solution.Secondary.Nudge)
	c.Secondary.applyAngularNudge(solution.Secondary.AngularNudge)
}

// JacobianImpulseLambda returns the impulse lambda for the specified
// constraint Jacobian, positional drift and restitution.
func (c DBSolverContext) JacobianImpulseLambda(jacobian PairJacobian, drift, restitution float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Primary, c.Secondary)
	if effMass < epsilon {
		return 0.0
	}
	effVelocity := jacobian.EffectiveVelocity(c.Primary, c.Secondary)
	resitutionClamp := RestitutionClamp(effVelocity)
	baumgarte := ImpulseDriftAdjustmentRatio * drift / c.ElapsedSeconds
	return -((1+restitution*resitutionClamp)*effVelocity + baumgarte) / effMass
}

// JacobianImpulseSolution returns an impulse solution based on the specified
// constraint Jacobian, positional drift and restitution.
func (c DBSolverContext) JacobianImpulseSolution(jacobian PairJacobian, drift, restitution float64) DBImpulseSolution {
	lambda := c.JacobianImpulseLambda(jacobian, drift, restitution)
	return jacobian.ImpulseSolution(lambda)
}

// JacobianNudgeLambda returns the nudge lambda for the specified
// constraint Jacobian and positional drift.
func (c DBSolverContext) JacobianNudgeLambda(jacobian PairJacobian, drift float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Primary, c.Secondary)
	if effMass < epsilon {
		return 0.0
	}
	return -NudgeDriftAdjustmentRatio * drift / effMass
}

// JacobianNudgeSolution returns a nudge solution based on the specified
// constraint Jacobian and positional drift.
func (c DBSolverContext) JacobianNudgeSolution(jacobian PairJacobian, drift float64) DBNudgeSolution {
	lambda := c.JacobianNudgeLambda(jacobian, drift)
	return jacobian.NudgeSolution(lambda)
}

// SBConstraintSolver represents the algorithm necessary
// to enforce a single-body constraint.
type SBConstraintSolver interface {
	// Reset clears the internal cache state for this constraint solver.
	// This is called at the start of every iteration.
	Reset(ctx SBSolverContext)

	// ApplyImpulses is called by the physics engine to instruct the solver
	// to apply the necessary impulses to its body.
	// This is called multiple times per iteration.
	ApplyImpulses(ctx SBSolverContext)

	// ApplyNudges is called by the physics engine to instruct the solver to
	// apply the necessary nudges to its body.
	// This is called multiple times per iteration.
	ApplyNudges(ctx SBSolverContext)
}

// DBConstraintSolver represents the algorithm necessary to enforce
// a double-body constraint.
type DBConstraintSolver interface {
	// Reset clears the internal cache state for this constraint solver.
	// This is called at the start of every iteration.
	Reset(ctx DBSolverContext)

	// ApplyImpulses is called by the physics engine to instruct the solver
	// to apply the necessary impulses to its bodies.
	// This is called multiple times per iteration.
	ApplyImpulses(ctx DBSolverContext)

	// ApplyNudges is called by the physics engine to instruct the solver to
	// apply the necessary nudges to its bodies.
	// This is called multiple times per iteration.
	ApplyNudges(ctx DBSolverContext)
}
