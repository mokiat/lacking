package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewStaticRotation creates a new StaticRotation constraint solver.
func NewStaticRotation() *StaticRotation {
	return &StaticRotation{
		rotation: dprec.IdentityQuat(),
	}
}

var _ physics.ExplicitSBConstraintSolver = (*StaticRotation)(nil)

// StaticRotation represents the solution for a constraint
// that keeps a body positioned at the specified fixture location.
//
// This solver is immediate - it does not use impulses or nudges.
type StaticRotation struct {
	physics.NilSBConstraintSolver // TODO: Remove

	rotation dprec.Quat
}

// Rotation returns the orientation to which the body will be constrained.
func (t *StaticRotation) Rotation() dprec.Quat {
	return t.rotation
}

// SetRotation changes the orientation to which the body will be constrained.
func (t *StaticRotation) SetRotation(rotation dprec.Quat) *StaticRotation {
	t.rotation = rotation
	return t
}

func (s *StaticRotation) ApplyImpulses(ctx physics.SBSolverContext) {
	ctx.Body.SetAngularVelocity(dprec.ZeroVec3())
}

func (s *StaticRotation) ApplyNudges(ctx physics.SBSolverContext) {
	ctx.Body.SetOrientation(s.rotation)
}
