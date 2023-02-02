package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewDifferential creates a new Differential constraint solver.
func NewDifferential() *Differential {
	return &Differential{
		maxDelta: 100.0,
	}
}

var _ solver.PairConstraint = (*Differential)(nil)

// Differential represents the solution for a constraint that keeps two
// objects from rotating too much relative to one another over the local X
// axis.
type Differential struct {
	maxDelta float64
}

// MaxDelta returns the maximum difference in velocity that is allowed.
func (d *Differential) MaxDelta() float64 {
	return d.maxDelta
}

// SetMaxDelta changes the maximum difference in velocity that is allowed.
func (d *Differential) SetMaxDelta(maxDelta float64) *Differential {
	d.maxDelta = maxDelta
	return d
}

func (d *Differential) Reset(ctx solver.PairContext) {}

func (d *Differential) ApplyImpulses(ctx solver.PairContext) {
	targetAxisX := ctx.Target.Rotation().OrientationX()
	targetVelocity := dprec.Vec3Dot(targetAxisX, ctx.Target.AngularVelocity())
	sourceAxisX := ctx.Source.Rotation().OrientationX()
	sourceVelocity := dprec.Vec3Dot(sourceAxisX, ctx.Source.AngularVelocity())

	var targetCorrection dprec.Vec3
	if delta := targetVelocity - sourceVelocity; delta > d.maxDelta {
		targetCorrection = dprec.Vec3Prod(targetAxisX, d.maxDelta-delta)
	}
	var sourceCorrectiom dprec.Vec3
	if delta := sourceVelocity - targetVelocity; delta > d.maxDelta {
		sourceCorrectiom = dprec.Vec3Prod(sourceAxisX, d.maxDelta-delta)
	}

	// FIXME: This is a naive implementation. It adjusts only the velocity
	// of one of the wheels, without affecting the other. Furthermore, it is
	// sharp when the threshold is passed, instead of gradual.
	ctx.Target.ApplyImpulse(solver.Impulse{
		Linear:  dprec.ZeroVec3(),
		Angular: targetCorrection,
	})
	ctx.Source.ApplyImpulse(solver.Impulse{
		Linear:  dprec.ZeroVec3(),
		Angular: sourceCorrectiom,
	})
}

func (d *Differential) ApplyNudges(ctx solver.PairContext) {}
