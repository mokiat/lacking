package solver

import "github.com/mokiat/gomath/dprec"

// PairJacobian represents the 1x12 Jacobian matrix of a double-object velocity
// constraint.
type PairJacobian struct {
	Target Jacobian
	Source Jacobian
}

// EffectiveVelocity returns the amount of the combined velocities of the two
// objects that is going in the wrong direction.
func (j PairJacobian) EffectiveVelocity(target, source *Placeholder) float64 {
	return j.Target.EffectiveVelocity(target) + j.Source.EffectiveVelocity(source)
}

// InverseEffectiveMass returns the inverse of the effective mass with which
// the two bodies affect the constraint.
func (j PairJacobian) InverseEffectiveMass(target, source *Placeholder) float64 {
	return j.Target.InverseEffectiveMass(target) + j.Source.InverseEffectiveMass(source)
}

// Impulse returns an impulse solution based on the lambda impulse
// amount applied according to this Jacobian.
func (j PairJacobian) Impulse(lambda float64) PairImpulse {
	return PairImpulse{
		Target: Impulse{
			Linear:  dprec.Vec3Prod(j.Target.LinearSlope, lambda),
			Angular: dprec.Vec3Prod(j.Target.AngularSlope, lambda),
		},
		Source: Impulse{
			Linear:  dprec.Vec3Prod(j.Source.LinearSlope, lambda),
			Angular: dprec.Vec3Prod(j.Source.AngularSlope, lambda),
		},
	}
}

// Nudge returns a nudge solution based on the lambda nudge amount
// applied according to this Jacobian.
func (j PairJacobian) Nudge(lambda float64) PairNudge {
	return PairNudge{
		Target: Nudge{
			Linear:  dprec.Vec3Prod(j.Target.LinearSlope, lambda),
			Angular: dprec.Vec3Prod(j.Target.AngularSlope, lambda),
		},
		Source: Nudge{
			Linear:  dprec.Vec3Prod(j.Source.LinearSlope, lambda),
			Angular: dprec.Vec3Prod(j.Source.AngularSlope, lambda),
		},
	}
}
