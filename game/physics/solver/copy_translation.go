package solver

import (
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*CopyTranslation)(nil)

// NewCopyTranslation creates a new CopyTranslation constraint solver.
func NewCopyTranslation() *CopyTranslation {
	return &CopyTranslation{}
}

type CopyTranslation struct {
	physics.NilDBConstraintSolver
}

func (s *CopyTranslation) CalculateNudges(ctx physics.DBSolverContext) physics.DBNudgeSolution {
	// TODO: Can we run this at the end?
	ctx.Secondary.SetPosition(ctx.Primary.Position())
	return physics.DBNudgeSolution{}
}
