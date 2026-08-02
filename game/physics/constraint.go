package physics

import "github.com/mokiat/gomath/dprec"

type Impulse struct {
	Linear  dprec.Vec3
	Angular dprec.Vec3
}

type Nudge struct {
	Linear  dprec.Vec3
	Angular dprec.Vec3
}

type ConstraintTarget struct {
	body *bodyState
}

func newConstraintTarget(body *bodyState) ConstraintTarget {
	return ConstraintTarget{
		body: body,
	}
}

func (t ConstraintTarget) InverseMass() float64 {
	return t.body.invMass
}

func (t ConstraintTarget) Mass() float64 {
	return t.body.mass()
}

func (t ConstraintTarget) InverseInertia() dprec.Mat3 {
	return t.body.invInertia
}

func (t ConstraintTarget) Inertia() dprec.Mat3 {
	return t.body.inertia()
}

func (t ConstraintTarget) LinearVelocity() dprec.Vec3 {
	return t.body.linearVelocity
}

func (t ConstraintTarget) SetLinearVelocity(velocity dprec.Vec3) {
	t.body.linearVelocity = velocity
}

func (t ConstraintTarget) AddLinearVelocity(delta dprec.Vec3) {
	t.body.addLinearVelocity(delta)
}

func (t ConstraintTarget) AngularVelocity() dprec.Vec3 {
	return t.body.angularVelocity
}

func (t ConstraintTarget) SetAngularVelocity(velocity dprec.Vec3) {
	t.body.angularVelocity = velocity
}

func (t ConstraintTarget) AddAngularVelocity(delta dprec.Vec3) {
	t.body.addAngularVelocity(delta)
}

func (t ConstraintTarget) ApplyImpulse(impulse Impulse) {
	t.body.addLinearVelocity(dprec.Vec3Prod(impulse.Linear, t.body.invMass))
	t.body.addAngularVelocity(dprec.Mat3Vec3Prod(t.body.invInertia, impulse.Angular))
}

func (t ConstraintTarget) Position() dprec.Vec3 {
	return t.body.position
}

func (t ConstraintTarget) SetPosition(position dprec.Vec3) {
	t.body.position = position
}

func (t ConstraintTarget) Translate(delta dprec.Vec3) {
	t.body.translate(delta)
}

func (t ConstraintTarget) Rotation() dprec.Quat {
	return t.body.rotation
}

func (t ConstraintTarget) SetRotation(rotation dprec.Quat) {
	t.body.rotation = rotation
}

func (t ConstraintTarget) Rotate(rotation dprec.Quat) {
	t.body.rotate(rotation)
}

func (t ConstraintTarget) ApplyNudge(nudge Nudge) {
	t.body.translate(dprec.Vec3Prod(nudge.Linear, t.body.invMass))
	t.body.rotate(QuatFromVector(dprec.Mat3Vec3Prod(t.body.invInertia, nudge.Angular)))
}

// Jacobian represents the 1x6 Jacobian matrix of a single-object velocity
// constraint.
type Jacobian struct {
	LinearSlope  dprec.Vec3
	AngularSlope dprec.Vec3
}

// EffectiveVelocity returns the amount of velocity in the wrong direction
// of the target.
func (j Jacobian) EffectiveVelocity(target ConstraintTarget) float64 {
	linear := dprec.Vec3Dot(j.LinearSlope, target.LinearVelocity())
	angular := dprec.Vec3Dot(j.AngularSlope, target.AngularVelocity())
	return linear + angular
}

// InverseEffectiveMass returns the inverse of the effective mass with which
// the target affects the constraint.
func (j Jacobian) InverseEffectiveMass(target ConstraintTarget) float64 {
	linear := dprec.Vec3Dot(j.LinearSlope, j.LinearSlope) * target.InverseMass()
	angular := dprec.Vec3Dot(dprec.Mat3Vec3Prod(target.InverseInertia(), j.AngularSlope), j.AngularSlope)
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
