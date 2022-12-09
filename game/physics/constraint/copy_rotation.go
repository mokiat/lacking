package constraint

import "github.com/mokiat/lacking/game/physics/solver"

// NewCopyRotation creates a new CopyRotation constraint solver.
func NewCopyRotation() *CopyRotation {
	return &CopyRotation{}
}

var _ solver.PairConstraint = (*CopyRotation)(nil)

// CopyRotation ensures that the target body has exactly the same rotation
// as the source one.
//
// This solver is immediate - it converges in a single step.
type CopyRotation struct{}

func (s *CopyRotation) Reset(ctx solver.PairContext) {}

func (s *CopyRotation) ApplyImpulses(ctx solver.PairContext) {
	ctx.Target.SetAngularVelocity(ctx.Source.AngularVelocity())
}

func (s *CopyRotation) ApplyNudges(ctx solver.PairContext) {
	ctx.Target.SetRotation(ctx.Source.Rotation())
}
