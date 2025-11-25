package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewPairAttachment creates a new PairAttachment constraint solver, which
// can be used to attach two bodies together at given offsets. Each body
// is still free to rotate independently.
func NewPairAttachment() *PairAttachment {
	solverX := NewMatchDirectionOffset().SetDirection(dprec.BasisXVec3()).SetOffset(0.0)
	solverY := NewMatchDirectionOffset().SetDirection(dprec.BasisYVec3()).SetOffset(0.0)
	solverZ := NewMatchDirectionOffset().SetDirection(dprec.BasisZVec3()).SetOffset(0.0)
	return &PairAttachment{
		solverX: *solverX,
		solverY: *solverY,
		solverZ: *solverZ,
	}
}

var _ solver.PairConstraint = (*PairAttachment)(nil)

// TODO: Implement the following constraint independently.

type PairAttachment struct {
	solverX MatchDirectionOffset
	solverY MatchDirectionOffset
	solverZ MatchDirectionOffset
}

func (s *PairAttachment) SetPrimaryOffset(offset dprec.Vec3) *PairAttachment {
	s.solverX.SetPrimaryRadius(offset)
	s.solverY.SetPrimaryRadius(offset)
	s.solverZ.SetPrimaryRadius(offset)
	return s
}

func (s *PairAttachment) SetSecondaryOffset(offset dprec.Vec3) *PairAttachment {
	s.solverX.SetSecondaryRadius(offset)
	s.solverY.SetSecondaryRadius(offset)
	s.solverZ.SetSecondaryRadius(offset)
	return s
}

func (s *PairAttachment) Reset(ctx solver.PairContext) {
	s.solverX.Reset(ctx)
	s.solverY.Reset(ctx)
	s.solverZ.Reset(ctx)
}

func (s *PairAttachment) ApplyImpulses(ctx solver.PairContext) {
	s.solverX.ApplyImpulses(ctx)
	s.solverY.ApplyImpulses(ctx)
	s.solverZ.ApplyImpulses(ctx)
}

func (s *PairAttachment) ApplyNudges(ctx solver.PairContext) {
	s.solverX.ApplyNudges(ctx)
	s.solverY.ApplyNudges(ctx)
	s.solverZ.ApplyNudges(ctx)
}
