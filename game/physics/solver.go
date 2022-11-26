package physics

import "github.com/mokiat/gomath/dprec"

// TODO: Nudges are slower because each nudge iteration moves the body
// a bit, causing it to be relocated in the octree. Either optimize
// octree relocations or better yet accumulate positional changes
// and apply them only at the end.
// This would mean that there needs to be a temporary replacement structure
// for a body that is used only during constraint evaluation.
// This might not be such a bad idea, since it could temporarily contain the
// inverse mass and moment of intertia, avoiding inverse matrix calculations.

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

// DBImpulseSolution is a solution to a constraint that
// indicates the impulses that need to be applied to the primary body
// and optionally (if the body is not nil) secondary body.
type DBImpulseSolution struct {
	Primary   SBImpulseSolution
	Secondary SBImpulseSolution
}

// DBNudgeSolution is a solution to a constraint that
// indicates the nudges that need to be applied to the primary body
// and optionally (if the body is not nil) secondary body.
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

// RestitutionCoefficient returns the restitution coefficient to be used for
// elastic collisions.
func (c SBSolverContext) RestitutionCoefficient() float64 {
	return c.Body.restitutionCoefficient
}

// ApplyImpulse is e helper function that applies an impulse to the body
// based on the specified jacobian.
func (c SBSolverContext) ApplyImpulse(jacobian Jacobian) {
	c.ApplyElasticImpulse(jacobian, 0.0)
}

// ApplyElasticImpulse is e helper function that applies an impulse to the body
// based on the specified jacobian and coefficient of restitution.
func (c SBSolverContext) ApplyElasticImpulse(jacobian Jacobian, restitution float64) {
	lambda := (1 + restitution) * jacobian.ImpulseLambda(c.Body)
	c.ApplyImpulseSolution(jacobian.ImpulseSolution(lambda))
}

// ApplyImpulseSolution applies the specified impulse solution to the relevant
// body.
func (c SBSolverContext) ApplyImpulseSolution(solution SBImpulseSolution) {
	c.Body.applyImpulse(solution.Impulse)
	c.Body.applyAngularImpulse(solution.AngularImpulse)
}

// ApplyNudge is e helper function that applies a nudge to the body
// based on the specified jacobian and positional drift.
func (c SBSolverContext) ApplyNudge(jacobian Jacobian, drift float64) {
	lambda := jacobian.NudgeLambda(c.Body, drift)
	c.ApplyNudgeSolution(jacobian.NudgeSolution(lambda))
}

// ApplyNudgeSolution applies the specified nudge solution to the relevant
// body.
func (c SBSolverContext) ApplyNudgeSolution(solution SBNudgeSolution) {
	c.Body.applyNudge(solution.Nudge)
	c.Body.applyAngularNudge(solution.AngularNudge)
}

// DBSolverContext contains information related to double-body constraint
// processing.
type DBSolverContext struct {
	Primary        *Body
	Secondary      *Body
	ElapsedSeconds float64
}

// RestitutionCoefficient returns the restitution coefficient to be used for
// elastic collisions.
func (c DBSolverContext) RestitutionCoefficient() float64 {
	return c.Primary.restitutionCoefficient * c.Secondary.restitutionCoefficient
}

// ApplyImpulse is e helper function that applies an impulse to the two bodies
// based on the specified jacobian.
func (c DBSolverContext) ApplyImpulse(jacobian PairJacobian) {
	c.ApplyElasticImpulse(jacobian, 0.0)
}

// ApplyElasticImpulse is e helper function that applies an impulse to the two
// bodies based on the specified jacobian and coefficient of restitution.
func (c DBSolverContext) ApplyElasticImpulse(jacobian PairJacobian, restitution float64) {
	lambda := (1 + restitution) * jacobian.ImpulseLambda(c.Primary, c.Secondary)
	c.ApplyImpulseSolution(jacobian.ImpulseSolution(lambda))
}

// ApplyImpulseSolution applies the specified impulse solution to the relevant
// bodies.
func (c DBSolverContext) ApplyImpulseSolution(solution DBImpulseSolution) {
	c.Primary.applyImpulse(solution.Primary.Impulse)
	c.Primary.applyAngularImpulse(solution.Primary.AngularImpulse)
	c.Secondary.applyImpulse(solution.Secondary.Impulse)
	c.Secondary.applyAngularImpulse(solution.Secondary.AngularImpulse)
}

// ApplyNudge is e helper function that applies a nudge to the two bodies
// based on the specified jacobian and positional drift.
func (c DBSolverContext) ApplyNudge(jacobian PairJacobian, drift float64) {
	lambda := jacobian.NudgeLambda(c.Primary, c.Secondary, drift)
	c.ApplyNudgeSolution(jacobian.NudgeSolution(lambda))
}

// ApplyNudgeSolution applies the specified nudge solution to the relevant
// bodies.
func (c DBSolverContext) ApplyNudgeSolution(solution DBNudgeSolution) {
	c.Primary.applyNudge(solution.Primary.Nudge)
	c.Primary.applyAngularNudge(solution.Primary.AngularNudge)
	c.Secondary.applyNudge(solution.Secondary.Nudge)
	c.Secondary.applyAngularNudge(solution.Secondary.AngularNudge)
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
