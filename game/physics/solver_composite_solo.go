package physics

// CompositeSoloConstraintSolver is a [SoloConstraintSolver] that combines
// several other solo constraint solvers into one, so that they can be
// registered as a single constraint through [SoloConstraintView.Create]
// - sharing one [SoloConstraintID], and consequently one enable/disable
// toggle - instead of one each.
//
// The composed solvers are driven in the order in which they were
// supplied to [NewCompositeSoloConstraintSolver]. Since they act on the
// same target one after another, rather than simultaneously, a solver
// later in the list observes any changes that an earlier one already
// made to the target during the same call.
//
// CompositeSoloConstraintSolver holds no configurable state of its own,
// so unlike most other solvers in this package, it has no Config type or
// Configure method; [NewCompositeSoloConstraintSolver] is the only way
// to obtain one.
type CompositeSoloConstraintSolver struct {
	solvers []SoloConstraintSolver
}

var _ SoloConstraintSolver = (*CompositeSoloConstraintSolver)(nil)

// NewCompositeSoloConstraintSolver creates a new
// [CompositeSoloConstraintSolver] that combines solvers, in the given
// order.
func NewCompositeSoloConstraintSolver(solvers ...SoloConstraintSolver) *CompositeSoloConstraintSolver {
	return &CompositeSoloConstraintSolver{
		solvers: solvers,
	}
}

// Solvers returns the solo constraint solvers that this solver combines,
// in the order in which they are driven.
func (s *CompositeSoloConstraintSolver) Solvers() []SoloConstraintSolver {
	return s.solvers
}

// Reset implements [SoloConstraintSolver.Reset].
//
// It calls [SoloConstraintSolver.Reset] on each combined solver, in
// order.
func (s *CompositeSoloConstraintSolver) Reset(ctx SoloConstraintContext) {
	for _, solver := range s.solvers {
		solver.Reset(ctx)
	}
}

// ApplyImpulses implements [SoloConstraintSolver.ApplyImpulses].
//
// It calls [SoloConstraintSolver.ApplyImpulses] on each combined solver,
// in order. This is safe without an intervening Reset, since impulses do
// not reposition the target and therefore cannot invalidate any
// position-derived state a combined solver cached during Reset.
func (s *CompositeSoloConstraintSolver) ApplyImpulses(ctx SoloConstraintContext) {
	for _, solver := range s.solvers {
		solver.ApplyImpulses(ctx)
	}
}

// ApplyNudges implements [SoloConstraintSolver.ApplyNudges].
//
// It calls [SoloConstraintSolver.ApplyNudges] on each combined solver, in
// order. A combined solver earlier in the list may reposition the
// target, but this requires no special handling here - per
// [SoloConstraintSolver.ApplyNudges], each combined solver is already
// responsible for recomputing any position-derived state it needs at the
// start of its own call.
func (s *CompositeSoloConstraintSolver) ApplyNudges(ctx SoloConstraintContext) {
	for _, solver := range s.solvers {
		solver.ApplyNudges(ctx)
	}
}
