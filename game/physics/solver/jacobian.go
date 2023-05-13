package solver

import "github.com/mokiat/gomath/dprec"

// Jacobian represents the 1x6 Jacobian matrix of a single-object velocity
// constraint.
type Jacobian struct {
	LinearSlope  dprec.Vec3
	AngularSlope dprec.Vec3
}

// EffectiveVelocity returns the amount of velocity in the wrong direction
// of the target.
func (j Jacobian) EffectiveVelocity(target *Placeholder) float64 {
	linear := dprec.Vec3Dot(j.LinearSlope, target.linearVelocity)
	angular := dprec.Vec3Dot(j.AngularSlope, target.angularVelocity)
	return linear + angular
}

// InverseEffectiveMass returns the inverse of the effective mass with which
// the target affects the constraint.
func (j Jacobian) InverseEffectiveMass(target *Placeholder) float64 {
	linear := dprec.Vec3Dot(j.LinearSlope, j.LinearSlope) * target.inverseMass
	angular := dprec.Vec3Dot(dprec.Mat3Vec3Prod(target.inverseMomentOfInertia, j.AngularSlope), j.AngularSlope)
	return linear + angular
}

// Impulse returns an Impulse solution based on the lambda impulse
// amount applied according to this Jacobian.
func (j Jacobian) Impulse(lambda float64) Impulse {
	return Impulse{
		Linear:  dprec.Vec3Prod(j.LinearSlope, lambda),
		Angular: dprec.Vec3Prod(j.AngularSlope, lambda),
	}
}

// Nudge returns a nudge solution based on the lambda nudge amount
// applied according to this Jacobian.
func (j Jacobian) Nudge(lambda float64) Nudge {
	return Nudge{
		Linear:  dprec.Vec3Prod(j.LinearSlope, lambda),
		Angular: dprec.Vec3Prod(j.AngularSlope, lambda),
	}
}

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
