package physics

import "github.com/mokiat/gomath/sprec"

// SBConstraintSolver represents the algorithm necessary
// to enforce a single-body constraint.
type SBConstraintSolver interface {

	// Reset clears the internal cache state for this constraint solver.
	// This is called at the start of every iteration.
	Reset()

	// CalculateImpulses returns a set of impulses to be applied
	// to the body.
	CalculateImpulses(body *Body, ctx ConstraintContext) SBImpulseSolution

	// CalculateNudges returns a set of nudges to be applied
	// to the body.
	CalculateNudges(body *Body, ctx ConstraintContext) SBNudgeSolution
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

// NilConstraintSolver is a ConstraintSolver that does nothing.
type NilSBConstraintSolver struct{}

func (s *NilSBConstraintSolver) Reset() {}

func (s *NilSBConstraintSolver) CalculateImpulses(body *Body, ctx ConstraintContext) SBImpulseSolution {
	return SBImpulseSolution{}
}

func (s *NilSBConstraintSolver) CalculateNudges(body *Body, ctx ConstraintContext) SBNudgeSolution {
	return SBNudgeSolution{}
}

var _ SBConstraintSolver = (*SBJacobianConstraintSolver)(nil)

// SBCalculateFunc is a function that calculates a jacobian for a
// single body constraint.
type SBCalculateFunc func(body *Body) (Jacobian, float32)

// NewSBJacobianConstraintSolver returns a new SBJacobianConstraintSolver
// based on the specified calculate function.
func NewSBJacobianConstraintSolver(calculate SBCalculateFunc) *SBJacobianConstraintSolver {
	return &SBJacobianConstraintSolver{
		calculate: calculate,
	}
}

// SBJacobianConstraintSolver is a helper implementation of
// SBConstraintSolver that is based on a Jacobian.
type SBJacobianConstraintSolver struct {
	calculate SBCalculateFunc
}

func (s *SBJacobianConstraintSolver) Reset() {}

func (s *SBJacobianConstraintSolver) CalculateImpulses(body *Body, ctx ConstraintContext) SBImpulseSolution {
	jacobian, drift := s.calculate(body)
	if sprec.Abs(drift) < epsilon {
		return SBImpulseSolution{}
	}
	lambda := jacobian.ImpulseLambda(body)
	return jacobian.ImpulseSolution(body, lambda)
}

func (s *SBJacobianConstraintSolver) CalculateNudges(body *Body, ctx ConstraintContext) SBNudgeSolution {
	jacobian, drift := s.calculate(body)
	if sprec.Abs(drift) < epsilon {
		return SBNudgeSolution{}
	}
	lambda := jacobian.NudgeLambda(body, drift)
	return jacobian.NudgeSolution(body, lambda)
}

// DBConstraintSolver represents the algorithm necessary to enforce
// a double-body constraint.
type DBConstraintSolver interface {

	// Reset clears the internal cache state for this constraint solver.
	// This is called at the start of every iteration.
	Reset()

	// CalculateImpulses returns a set of impulses to be applied
	// to the two bodies.
	CalculateImpulses(primary, secondary *Body, ctx ConstraintContext) DBImpulseSolution

	// CalculateNudges returns a set of nudges to be applied
	// to the two bodies.
	CalculateNudges(primary, secondary *Body, ctx ConstraintContext) DBNudgeSolution
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

// NilConstraintSolver is a ConstraintSolver that does nothing.
type NilDBConstraintSolver struct{}

func (s *NilDBConstraintSolver) Reset() {}

func (s *NilDBConstraintSolver) CalculateImpulses(primary, secondary *Body, ctx ConstraintContext) DBImpulseSolution {
	return DBImpulseSolution{}
}

func (s *NilDBConstraintSolver) CalculateNudges(primary, secondary *Body, ctx ConstraintContext) DBNudgeSolution {
	return DBNudgeSolution{}
}
