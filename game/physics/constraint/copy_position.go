package constraint

import "github.com/mokiat/lacking/game/physics/solver"

// NewCopyPosition creates a new CopyPosition constraint solver.
func NewCopyPosition() *CopyPosition {
	return &CopyPosition{}
}

var _ solver.PairConstraint = (*CopyPosition)(nil)

// CopyPosition ensures that the target object has the same position as
// the source one.
//
// This solver is immediate - it converges in a single step.
type CopyPosition struct{}

func (s *CopyPosition) Reset(ctx solver.PairContext) {}

func (s *CopyPosition) ApplyImpulses(ctx solver.PairContext) {
	ctx.Target.SetLinearVelocity(ctx.Source.LinearVelocity())
}

func (s *CopyPosition) ApplyNudges(ctx solver.PairContext) {
	ctx.Target.SetPosition(ctx.Source.Position())
}
