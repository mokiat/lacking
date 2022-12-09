package constraint

import "github.com/mokiat/lacking/game/physics/solver"

// NewCombined creates a new Combined solver based on the specified
// sub-solvers.
func NewCombined(delegates ...solver.Constraint) *Combined {
	return &Combined{
		delegates: delegates,
	}
}

var _ solver.Constraint = (*Combined)(nil)

// Combined is a single-object solver that delegates its logic to a
// number of sub-solvers.
type Combined struct {
	delegates []solver.Constraint
}

func (s *Combined) Reset(ctx solver.Context) {
	for _, delegate := range s.delegates {
		delegate.Reset(ctx)
	}
}

func (s *Combined) ApplyImpulses(ctx solver.Context) {
	for _, delegate := range s.delegates {
		delegate.ApplyImpulses(ctx)
	}
}

func (s *Combined) ApplyNudges(ctx solver.Context) {
	for _, delegate := range s.delegates {
		delegate.ApplyNudges(ctx)
	}
}

// NewPairCombined creates a new PairCombined solver based on the specified
// sub-solvers.
func NewPairCombined(delegates ...solver.PairConstraint) *PairCombined {
	return &PairCombined{
		delegates: delegates,
	}
}

var _ solver.PairConstraint = (*PairCombined)(nil)

// PairCombined is a double-object solver that delegates its logic to a
// number of sub-solvers.
type PairCombined struct {
	delegates []solver.PairConstraint
}

func (s *PairCombined) Reset(ctx solver.PairContext) {
	for _, delegate := range s.delegates {
		delegate.Reset(ctx)
	}
}

func (s *PairCombined) ApplyImpulses(ctx solver.PairContext) {
	for _, delegate := range s.delegates {
		delegate.ApplyImpulses(ctx)
	}
}

func (s *PairCombined) ApplyNudges(ctx solver.PairContext) {
	for _, delegate := range s.delegates {
		delegate.ApplyNudges(ctx)
	}
}
