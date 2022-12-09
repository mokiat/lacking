package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewStaticPosition creates a new StaticPosition constraint solver.
func NewStaticPosition() *StaticPosition {
	return &StaticPosition{
		position: dprec.ZeroVec3(),
	}
}

var _ solver.Constraint = (*StaticPosition)(nil)

// StaticPosition represents the solution for a constraint
// that keeps a body positioned at the specified fixture location.
//
// This solver is immediate - it converges in a single step.
type StaticPosition struct {
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

func (s *StaticPosition) Reset(ctx solver.Context) {}

func (s *StaticPosition) ApplyImpulses(ctx solver.Context) {
	ctx.Target.SetLinearVelocity(dprec.ZeroVec3())
}

func (s *StaticPosition) ApplyNudges(ctx solver.Context) {
	ctx.Target.SetPosition(s.position)
}
