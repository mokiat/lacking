package solver

import "github.com/mokiat/lacking/game/physics"

// NewSBCombined creates a new SBCombined solver based on the specified
// sub-solvers.
func NewSBCombined(delegates ...physics.SBConstraintSolver) *SBCombined {
	return &SBCombined{
		delegates: delegates,
	}
}

var _ physics.SBConstraintSolver = (*SBCombined)(nil)

// SBCombined is a single-body solver that delegates its logic to a
// number of sub-solvers.
type SBCombined struct {
	delegates []physics.SBConstraintSolver
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
func NewDBCombined(delegates ...physics.DBConstraintSolver) *DBCombined {
	return &DBCombined{
		delegates: delegates,
	}
}

var _ physics.DBConstraintSolver = (*DBCombined)(nil)

// DBCombined is a double-body solver that delegates its logic to a
// number of sub-solvers.
type DBCombined struct {
	delegates []physics.DBConstraintSolver
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
