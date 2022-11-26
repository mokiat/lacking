package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewCoilover creates a new Coilover constraint solver.
func NewCoilover() *Coilover {
	return &Coilover{
		primaryRadius:   dprec.ZeroVec3(),
		secondaryRadius: dprec.ZeroVec3(),
		frequency:       1.0,
		damping:         0.5,
	}
}

var _ physics.DBConstraintSolver = (*Coilover)(nil)

// Coilover represents the solution for a constraint that immitates
// a car coilover through a damped harmonic oscillator.
type Coilover struct {
	primaryRadius   dprec.Vec3
	secondaryRadius dprec.Vec3
	frequency       float64
	damping         float64

	appliedLambda float64
	jacobian      physics.PairJacobian
}

// PrimaryRadius returns the radius vector of the contact point
// on the primary object.
//
// The vector is in the object's local space.
func (s *Coilover) PrimaryRadius() dprec.Vec3 {
	return s.primaryRadius
}

// SetPrimaryRadius changes the radius vector of the contact point
// on the primary object.
//
// The vector is in the object's local space.
func (s *Coilover) SetPrimaryRadius(radius dprec.Vec3) *Coilover {
	s.primaryRadius = radius
	return s
}

// SecondaryRadius returns the radius vector of the contact point
// on the secondary object.
//
// The vector is in the object's local space.
func (s *Coilover) SecondaryRadius() dprec.Vec3 {
	return s.secondaryRadius
}

// SetSecondaryRadius changes the radius vector of the contact point
// on the secondary object.
//
// The vector is in the object's local space.
func (s *Coilover) SetSecondaryRadius(radius dprec.Vec3) *Coilover {
	s.secondaryRadius = radius
	return s
}

// Frequency returns the frequency (in Hz) of the damped harmonic
// oscillator that represents this coilover.
func (s *Coilover) Frequency() float64 {
	return s.frequency
}

// SetFrequency changes the frequency (in Hz) of the damped harmonic
// oscillator that represents this coilover.
func (s *Coilover) SetFrequency(frequency float64) *Coilover {
	s.frequency = frequency
	return s
}

// Damping returns the damping ratio of the damped harmonic oscillator
// that represents this coilover.
func (s *Coilover) Damping() float64 {
	return s.damping
}

// SetDamping changes the damping ratio of the damped harmonic oscillator
// that represents this coilover.
func (s *Coilover) SetDamping(damping float64) *Coilover {
	s.damping = damping
	return s
}

func (s *Coilover) Reset(physics.DBSolverContext) {
	s.appliedLambda = 0.0
}

func (s *Coilover) ApplyImpulses(ctx physics.DBSolverContext) {
	if dprec.Abs(ctx.ElapsedSeconds) < epsilon {
		return
	}

	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.primaryRadius)
	primaryPointWS := dprec.Vec3Sum(ctx.Primary.Position(), primaryRadiusWS)
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), s.secondaryRadius)
	secondaryPointWS := dprec.Vec3Sum(ctx.Secondary.Position(), secondaryRadiusWS)

	deltaPosition := dprec.Vec3Diff(secondaryPointWS, primaryPointWS)
	drift := deltaPosition.Length()
	if drift < epsilon {
		return
	}
	normal := dprec.UnitVec3(deltaPosition)

	jacobian := physics.PairJacobian{
		Primary: physics.Jacobian{
			SlopeVelocity:        dprec.InverseVec3(normal),
			SlopeAngularVelocity: dprec.Vec3Cross(normal, primaryRadiusWS),
		},
		Secondary: physics.Jacobian{
			SlopeVelocity:        normal,
			SlopeAngularVelocity: dprec.Vec3Cross(secondaryRadiusWS, normal),
		},
	}

	invertedEffectiveMass := jacobian.InverseEffectiveMass(ctx.Primary, ctx.Secondary)
	w := 2.0 * dprec.Pi * s.frequency
	dc := 2.0 * s.damping * w / invertedEffectiveMass
	k := w * w / invertedEffectiveMass

	gamma := 1.0 / (ctx.ElapsedSeconds * (dc + ctx.ElapsedSeconds*k))
	beta := ctx.ElapsedSeconds * k * gamma

	effectiveVelocity := jacobian.EffectiveVelocity(ctx.Primary, ctx.Secondary)
	lambda := -(effectiveVelocity + beta*drift + gamma*s.appliedLambda) / (invertedEffectiveMass + gamma)
	ctx.ApplyImpulseSolution(jacobian.ImpulseSolution(lambda))
	s.appliedLambda += lambda
}

func (s *Coilover) ApplyNudges(ctx physics.DBSolverContext) {
	// no nudge corrections for coilover
}
