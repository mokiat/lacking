package physics

// CompositePairConstraintSolver is a [PairConstraintSolver] that combines
// several other pair constraint solvers into one, so that they can be
// registered as a single constraint through [PairConstraintView.Create]
// - sharing one [PairConstraintID], and consequently one enable/disable
// toggle - instead of one each.
//
// The composed solvers are driven in the order in which they were
// supplied to [NewCompositePairConstraintSolver]. Since they act on the
// same targets one after another, rather than simultaneously, a solver
// later in the list observes any changes that an earlier one already
// made to the targets during the same call.
//
// CompositePairConstraintSolver holds no configurable state of its own,
// so unlike most other solvers in this package, it has no Config type or
// Configure method; [NewCompositePairConstraintSolver] is the only way
// to obtain one.
type CompositePairConstraintSolver struct {
	solvers []PairConstraintSolver
}

var _ PairConstraintSolver = (*CompositePairConstraintSolver)(nil)

// NewCompositePairConstraintSolver creates a new
// [CompositePairConstraintSolver] that combines solvers, in the given
// order.
func NewCompositePairConstraintSolver(solvers ...PairConstraintSolver) *CompositePairConstraintSolver {
	return &CompositePairConstraintSolver{
		solvers: solvers,
	}
}

// Solvers returns the pair constraint solvers that this solver combines,
// in the order in which they are driven.
func (s *CompositePairConstraintSolver) Solvers() []PairConstraintSolver {
	return s.solvers
}

// Reset implements [PairConstraintSolver.Reset].
//
// It calls [PairConstraintSolver.Reset] on each combined solver, in
// order.
func (s *CompositePairConstraintSolver) Reset(ctx PairConstraintContext) {
	for _, solver := range s.solvers {
		solver.Reset(ctx)
	}
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It calls [PairConstraintSolver.ApplyImpulses] on each combined solver,
// in order. This is safe without an intervening Reset, since impulses do
// not reposition the targets and therefore cannot invalidate any
// position-derived state a combined solver cached during Reset.
func (s *CompositePairConstraintSolver) ApplyImpulses(ctx PairConstraintContext) {
	for _, solver := range s.solvers {
		solver.ApplyImpulses(ctx)
	}
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It calls [PairConstraintSolver.Reset] followed by
// [PairConstraintSolver.ApplyNudges] on each combined solver in turn,
// rather than calling ApplyNudges on all of them in a single pass. This
// preserves the guarantee, documented on [PairConstraintSolver.Reset],
// that Reset always immediately precedes ApplyNudges for a given solver
// - since a combined solver earlier in the list may reposition either
// target, which would otherwise leave a later solver's cached,
// position-derived state stale for the remainder of this call.
func (s *CompositePairConstraintSolver) ApplyNudges(ctx PairConstraintContext) {
	for _, solver := range s.solvers {
		solver.Reset(ctx) // preserve engine reset behavior
		solver.ApplyNudges(ctx)
	}
}
