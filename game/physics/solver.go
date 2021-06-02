package physics

import "github.com/mokiat/gomath/sprec"

// SBSolverContext contains information related to the
// single-body constraint processing.
type SBSolverContext struct {
	Body           *Body
	ElapsedSeconds float32
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

// SBImpulseSolution is a solution to a single-body constraint that
// contains the impulses that need to be applied to the body.
type SBImpulseSolution struct {
	Impulse        sprec.Vec3
	AngularImpulse sprec.Vec3
}

// SBNudgeSolution is a solution to a single-body constraint that
// contains the nudges that need to be applied to the body.
type SBNudgeSolution struct {
	Nudge        sprec.Vec3
	AngularNudge sprec.Vec3
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
type SBCalculateFunc func(SBSolverContext) (Jacobian, float32)

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
	if sprec.Abs(drift) < epsilon {
		return SBImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(ctx.Body)
	return jacobian.ImpulseSolution(ctx.Body, lambda)
}

func (s *SBJacobianConstraintSolver) CalculateNudges(ctx SBSolverContext) SBNudgeSolution {
	jacobian, drift := s.calculate(ctx)
	if sprec.Abs(drift) < epsilon {
		return SBNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(ctx.Body, drift)
	return jacobian.NudgeSolution(ctx.Body, lambda)
}

// DBSolverContext contains information related to the
// double-body constraint processing.
type DBSolverContext struct {
	Primary        *Body
	Secondary      *Body
	ElapsedSeconds float32
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
type DBCalculateFunc func(DBSolverContext) (PairJacobian, float32)

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
	if sprec.Abs(drift) < epsilon {
		return DBImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(ctx.Primary, ctx.Secondary)
	return jacobian.ImpulseSolution(ctx.Primary, ctx.Secondary, lambda)
}

func (s *DBJacobianConstraintSolver) CalculateNudges(ctx DBSolverContext) DBNudgeSolution {
	jacobian, drift := s.calculate(ctx)
	if sprec.Abs(drift) < epsilon {
		return DBNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(ctx.Primary, ctx.Secondary, drift)
	return jacobian.NudgeSolution(ctx.Primary, ctx.Secondary, lambda)
}
