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

	// CalculateImpulses returns a set of impulses to be applied
	// to the body.
	CalculateImpulses(ctx SBSolverContext) SBImpulseSolution

	// CalculateNudges returns a set of nudges to be applied
	// to the body.
	CalculateNudges(ctx SBSolverContext) SBNudgeSolution
}

// TODO: Rename
type ExplicitSBConstraintSolver interface {
	SBConstraintSolver // TODO: Remove

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

// TODO: Rename
type ExplicitDBConstraintSolver interface {
	DBConstraintSolver // TODO: Remove

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

var _ SBConstraintSolver = (*NilSBConstraintSolver)(nil)

// NilConstraintSolver is an SBConstraintSolver that does nothing.
type NilSBConstraintSolver struct{}

func (s *NilSBConstraintSolver) Reset(SBSolverContext) {}

func (s *NilSBConstraintSolver) CalculateImpulses(SBSolverContext) SBImpulseSolution {
	return SBImpulseSolution{}
}

func (s *NilSBConstraintSolver) CalculateNudges(SBSolverContext) SBNudgeSolution {
	return SBNudgeSolution{}
}

// SBCalculateFunc is a function that calculates a jacobian for a
// single body constraint.
type SBCalculateFunc func(SBSolverContext) (Jacobian, float64)

// NewSBJacobianConstraintSolver returns a new SBJacobianConstraintSolver
// based on the specified calculate function.
func NewSBJacobianConstraintSolver(calculate SBCalculateFunc) *SBJacobianConstraintSolver {
	return &SBJacobianConstraintSolver{
		calculate: calculate,
	}
}

var _ SBConstraintSolver = (*SBJacobianConstraintSolver)(nil)

// SBJacobianConstraintSolver is a helper implementation of
// SBConstraintSolver that is based on a Jacobian.
type SBJacobianConstraintSolver struct {
	calculate SBCalculateFunc
}

func (s *SBJacobianConstraintSolver) Reset(SBSolverContext) {}

func (s *SBJacobianConstraintSolver) CalculateImpulses(ctx SBSolverContext) SBImpulseSolution {
	jacobian, drift := s.calculate(ctx)
	// FIXME: Ignore drift!
	if dprec.Abs(drift) < epsilon {
		return SBImpulseSolution{}
	}
	return jacobian.ImpulseSolution(jacobian.ImpulseLambda(ctx.Body))
}

func (s *SBJacobianConstraintSolver) CalculateNudges(ctx SBSolverContext) SBNudgeSolution {
	jacobian, drift := s.calculate(ctx)
	// FIXME: Try without this?
	if dprec.Abs(drift) < epsilon {
		return SBNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(ctx.Body, drift)
	return jacobian.NudgeSolution(lambda)
}

// DBConstraintSolver represents the algorithm necessary to enforce
// a double-body constraint.
type DBConstraintSolver interface {

	// Reset clears the internal cache state for this constraint solver.
	// This is called at the start of every iteration.
	Reset(ctx DBSolverContext)

	// CalculateImpulses returns a set of impulses to be applied
	// to the two bodies.
	CalculateImpulses(ctx DBSolverContext) DBImpulseSolution

	// CalculateNudges returns a set of nudges to be applied
	// to the two bodies.
	CalculateNudges(ctx DBSolverContext) DBNudgeSolution
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

var _ DBConstraintSolver = (*NilDBConstraintSolver)(nil)

// NilConstraintSolver is a DBConstraintSolver that does nothing.
type NilDBConstraintSolver struct{}

func (s *NilDBConstraintSolver) Reset(DBSolverContext) {}

func (s *NilDBConstraintSolver) CalculateImpulses(DBSolverContext) DBImpulseSolution {
	return DBImpulseSolution{}
}

func (s *NilDBConstraintSolver) CalculateNudges(DBSolverContext) DBNudgeSolution {
	return DBNudgeSolution{}
}

// DBCalculateFunc is a function that calculates a jacobian for a
// double body constraint.
type DBCalculateFunc func(DBSolverContext) (PairJacobian, float64)

// NewDBJacobianConstraintSolver returns a new DBJacobianConstraintSolver
// based on the specified calculate function.
func NewDBJacobianConstraintSolver(calculate DBCalculateFunc) *DBJacobianConstraintSolver {
	return &DBJacobianConstraintSolver{
		calculate: calculate,
	}
}

var _ DBConstraintSolver = (*DBJacobianConstraintSolver)(nil)

// DBJacobianConstraintSolver is a helper implementation of
// DBConstraintSolver that is based on a Jacobian.
type DBJacobianConstraintSolver struct {
	calculate DBCalculateFunc
}

func (s *DBJacobianConstraintSolver) Reset(DBSolverContext) {}

func (s *DBJacobianConstraintSolver) CalculateImpulses(ctx DBSolverContext) DBImpulseSolution {
	jacobian, drift := s.calculate(ctx)
	// FIXME: Ignore drift!
	if dprec.Abs(drift) < epsilon {
		return DBImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(ctx.Primary, ctx.Secondary)
	return jacobian.ImpulseSolution(lambda)
}

func (s *DBJacobianConstraintSolver) CalculateNudges(ctx DBSolverContext) DBNudgeSolution {
	jacobian, drift := s.calculate(ctx)
	// FIXME: Try without this?
	if dprec.Abs(drift) < epsilon {
		return DBNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(ctx.Primary, ctx.Secondary, drift)
	return jacobian.NudgeSolution(lambda)
}
