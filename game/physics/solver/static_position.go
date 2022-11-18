package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewStaticPosition creates a new StaticPosition constraint solver.
func NewStaticPosition() *StaticPosition {
	return &StaticPosition{
		position: dprec.ZeroVec3(),
	}
}

var _ physics.ExplicitSBConstraintSolver = (*StaticPosition)(nil)

// StaticPosition represents the solution for a constraint
// that keeps a body positioned at the specified fixture location.
//
// This solver is immediate - it does not use impulses or nudges.
type StaticPosition struct {
	physics.NilSBConstraintSolver // TODO: Remove

	position dprec.Vec3
}

// Position returns the location to which the body will be constrained.
func (t *StaticPosition) Position() dprec.Vec3 {
	return t.position
}

// SetPosition changes the location to which the body will be constrained.
func (t *StaticPosition) SetPosition(position dprec.Vec3) *StaticPosition {
	t.position = position
	return t
}

func (s *StaticPosition) ApplyImpulses(ctx physics.SBSolverContext) {
	ctx.Body.SetVelocity(dprec.ZeroVec3())
}

func (s *StaticPosition) ApplyNudges(ctx physics.SBSolverContext) {
	ctx.Body.SetPosition(s.position)
}
