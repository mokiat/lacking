package solver

import (
	"github.com/mokiat/lacking/game/physics"
)

// NewCopyRotation creates a new CopyRotation constraint solver.
func NewCopyRotation() *CopyRotation {
	return &CopyRotation{}
}

var _ physics.DBConstraintSolver = (*CopyRotation)(nil)

// CopyRotation ensures that the second body has exactly the same orientation
// as the first one.
type CopyRotation struct{}

func (s *CopyRotation) Reset(ctx physics.DBSolverContext) {}

func (s *CopyRotation) ApplyImpulses(ctx physics.DBSolverContext) {
	ctx.Secondary.SetAngularVelocity(ctx.Primary.AngularVelocity())
}

func (s *CopyRotation) ApplyNudges(ctx physics.DBSolverContext) {
	ctx.Secondary.SetOrientation(ctx.Primary.Orientation())
}
