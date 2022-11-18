package solver

import "github.com/mokiat/lacking/game/physics"

// NewSBCombined creates a new SBCombined solver based on the specified
// sub-solvers.
func NewSBCombined(delegates ...physics.ExplicitSBConstraintSolver) *SBCombined {
	return &SBCombined{
		delegates: delegates,
	}
}

var _ physics.ExplicitSBConstraintSolver = (*SBCombined)(nil)

// SBCombined is a single-body solver that delegates its logic to a
// number of sub-solvers.
type SBCombined struct {
	physics.NilSBConstraintSolver // TODO: Remove

	delegates []physics.ExplicitSBConstraintSolver
}

func (s *SBCombined) Reset(ctx physics.SBSolverContext) {
	for _, delegate := range s.delegates {
		delegate.Reset(ctx)
	}
}

func (s *SBCombined) ApplyImpulses(ctx physics.SBSolverContext) {
	for _, delegate := range s.delegates {
		delegate.ApplyImpulses(ctx)
	}
}

func (s *SBCombined) ApplyNudges(ctx physics.SBSolverContext) {
	for _, delegate := range s.delegates {
		delegate.ApplyNudges(ctx)
	}
}

// NewDBCombined creates a new DBCombined solver based on the specified
// sub-solvers.
func NewDBCombined(delegates ...physics.ExplicitDBConstraintSolver) *DBCombined {
	return &DBCombined{
		delegates: delegates,
	}
}

var _ physics.ExplicitDBConstraintSolver = (*DBCombined)(nil)

// DBCombined is a double-body solver that delegates its logic to a
// number of sub-solvers.
type DBCombined struct {
	physics.NilDBConstraintSolver // TODO: Remove

	delegates []physics.ExplicitDBConstraintSolver
}

func (s *DBCombined) Reset(ctx physics.DBSolverContext) {
	for _, delegate := range s.delegates {
		delegate.Reset(ctx)
	}
}

func (s *DBCombined) ApplyImpulses(ctx physics.DBSolverContext) {
	for _, delegate := range s.delegates {
		delegate.ApplyImpulses(ctx)
	}
}

func (s *DBCombined) ApplyNudges(ctx physics.DBSolverContext) {
	for _, delegate := range s.delegates {
		delegate.ApplyNudges(ctx)
	}
}
