package solver

import (
	"github.com/mokiat/lacking/game/physics"
)

// NewCopyTranslation creates a new CopyTranslation constraint solver.
func NewCopyTranslation() *CopyTranslation {
	return &CopyTranslation{}
}

var _ physics.ExplicitDBConstraintSolver = (*CopyTranslation)(nil)

// CopyTranslation ensures that the second body has the same translation as
// the first one.
// This solver is immediate - it does not use impulses or nudges.
type CopyTranslation struct {
	physics.NilDBConstraintSolver // TODO: Remove
}

func (s *CopyTranslation) ApplyImpulses(ctx physics.DBSolverContext) {
	ctx.Secondary.SetVelocity(ctx.Primary.Velocity())
}

func (s *CopyTranslation) ApplyNudges(ctx physics.DBSolverContext) {
	ctx.Secondary.SetPosition(ctx.Primary.Position())
}
