package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewStaticRotation creates a new StaticRotation constraint solver.
func NewStaticRotation() *StaticRotation {
	return &StaticRotation{
		rotation: dprec.IdentityQuat(),
	}
}

var _ solver.Constraint = (*StaticRotation)(nil)

// StaticRotation represents the solution for a constraint
// that keeps a body positioned at the specified fixture location.
//
// This solver is immediate - it converges in a single step.
type StaticRotation struct {
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

func (s *StaticRotation) Reset(ctx solver.Context) {}

func (s *StaticRotation) ApplyImpulses(ctx solver.Context) {
	ctx.Target.SetAngularVelocity(dprec.ZeroVec3())
}

func (s *StaticRotation) ApplyNudges(ctx solver.Context) {
	ctx.Target.SetRotation(s.rotation)
}
