package physics

import "github.com/mokiat/gomath/dprec"

// Impulse represents a velocity-space correction, split into a linear and
// an angular component, that can be applied to a [ConstraintTarget]
// through [ConstraintTarget.ApplyImpulse].
type Impulse struct {

	// Linear is the change in linear velocity to be applied.
	Linear dprec.Vec3

	// Angular is the change in angular velocity to be applied.
	Angular dprec.Vec3
}

// Nudge represents a position-space correction, split into a linear and
// an angular component, that can be applied to a [ConstraintTarget]
// through [ConstraintTarget.ApplyNudge].
//
// It mirrors [Impulse], but acts on position and rotation directly
// instead of on velocity.
type Nudge struct {

	// Linear is the positional offset to be applied.
	Linear dprec.Vec3

	// Angular is the rotational offset to be applied, expressed as a
	// scaled rotation vector (see [QuatFromVector]).
	Angular dprec.Vec3
}

// ConstraintTarget represents the body acted upon by a
// [SoloConstraintSolver] or [PairConstraintSolver]. It exposes the
// body's motion state to the solver and lets the solver correct that
// state through impulses and nudges.
type ConstraintTarget struct {
	body *bodyState
}

// newConstraintTarget creates a new ConstraintTarget that wraps body.
func newConstraintTarget(body *bodyState) ConstraintTarget {
	return ConstraintTarget{
		body: body,
	}
}

// InverseMass returns the inverse of the target's mass.
func (t ConstraintTarget) InverseMass() float64 {
	return t.body.invMass
}

// Mass returns the target's mass.
func (t ConstraintTarget) Mass() float64 {
	return t.body.mass()
}

// InverseInertia returns the inverse of the target's world-space
// inertia tensor.
func (t ConstraintTarget) InverseInertia() dprec.Mat3 {
	return t.body.invInertia
}

// Inertia returns the target's world-space inertia tensor.
func (t ConstraintTarget) Inertia() dprec.Mat3 {
	return t.body.inertia()
}

// LinearVelocity returns the target's current linear velocity.
func (t ConstraintTarget) LinearVelocity() dprec.Vec3 {
	return t.body.linearVelocity
}

// SetLinearVelocity changes the target's linear velocity.
func (t ConstraintTarget) SetLinearVelocity(velocity dprec.Vec3) {
	t.body.linearVelocity = velocity
}

// AddLinearVelocity adds delta to the target's linear velocity.
func (t ConstraintTarget) AddLinearVelocity(delta dprec.Vec3) {
	t.body.addLinearVelocity(delta)
}

// AngularVelocity returns the target's current angular velocity.
func (t ConstraintTarget) AngularVelocity() dprec.Vec3 {
	return t.body.angularVelocity
}

// SetAngularVelocity changes the target's angular velocity.
func (t ConstraintTarget) SetAngularVelocity(velocity dprec.Vec3) {
	t.body.angularVelocity = velocity
}

// AddAngularVelocity adds delta to the target's angular velocity.
func (t ConstraintTarget) AddAngularVelocity(delta dprec.Vec3) {
	t.body.addAngularVelocity(delta)
}

// ApplyImpulse adjusts the target's linear and angular velocity
// according to impulse, scaled by the target's inverse mass and
// inverse inertia respectively.
func (t ConstraintTarget) ApplyImpulse(impulse Impulse) {
	t.body.addLinearVelocity(dprec.Vec3Prod(impulse.Linear, t.body.invMass))
	t.body.addAngularVelocity(dprec.Mat3Vec3Prod(t.body.invInertia, impulse.Angular))
}

// Position returns the target's current position.
func (t ConstraintTarget) Position() dprec.Vec3 {
	return t.body.position
}

// SetPosition changes the target's position.
func (t ConstraintTarget) SetPosition(position dprec.Vec3) {
	t.body.position = position
}

// Translate offsets the target's position by delta.
func (t ConstraintTarget) Translate(delta dprec.Vec3) {
	t.body.translate(delta)
}

// Rotation returns the target's current rotation.
func (t ConstraintTarget) Rotation() dprec.Quat {
	return t.body.rotation
}

// SetRotation changes the target's rotation.
func (t ConstraintTarget) SetRotation(rotation dprec.Quat) {
	t.body.rotation = dprec.UnitQuat(rotation)
}

// Rotate applies rotation on top of the target's current rotation.
func (t ConstraintTarget) Rotate(rotation dprec.Quat) {
	t.body.rotate(rotation)
}

// ApplyNudge adjusts the target's position and rotation according to
// nudge, scaled by the target's inverse mass and inverse inertia
// respectively, the same way [ConstraintTarget.ApplyImpulse] adjusts
// velocity.
func (t ConstraintTarget) ApplyNudge(nudge Nudge) {
	t.body.translate(dprec.Vec3Prod(nudge.Linear, t.body.invMass))
	t.body.rotate(QuatFromVector(dprec.Mat3Vec3Prod(t.body.invInertia, nudge.Angular)))
}

// Jacobian represents the 1x6 Jacobian matrix of a single-object velocity
// constraint, split into the 1x3 blocks that act on linear and angular
// velocity respectively.
type Jacobian struct {

	// LinearSlope is the block of the Jacobian that acts on linear
	// velocity.
	LinearSlope dprec.Vec3

	// AngularSlope is the block of the Jacobian that acts on angular
	// velocity.
	AngularSlope dprec.Vec3
}

// EffectiveVelocity returns the amount of velocity in the wrong direction
// of the constraint, i.e. the constraint equation's velocity error,
// evaluated against target's current linear and angular velocity.
func (j Jacobian) EffectiveVelocity(target ConstraintTarget) float64 {
	linear := dprec.Vec3Dot(j.LinearSlope, target.LinearVelocity())
	angular := dprec.Vec3Dot(j.AngularSlope, target.AngularVelocity())
	return linear + angular
}

// InverseEffectiveMass returns the inverse of the effective mass with which
// target affects the constraint (i.e. J * M^-1 * J^T, where J is this
// Jacobian and M is target's mass-inertia matrix).
func (j Jacobian) InverseEffectiveMass(target ConstraintTarget) float64 {
	linear := dprec.Vec3Dot(j.LinearSlope, j.LinearSlope) * target.InverseMass()
	angular := dprec.Vec3Dot(dprec.Mat3Vec3Prod(target.InverseInertia(), j.AngularSlope), j.AngularSlope)
	return linear + angular
}

// Impulse returns an [Impulse] solution based on the lambda impulse
// amount applied according to this Jacobian.
func (j Jacobian) Impulse(lambda float64) Impulse {
	return Impulse{
		Linear:  dprec.Vec3Prod(j.LinearSlope, lambda),
		Angular: dprec.Vec3Prod(j.AngularSlope, lambda),
	}
}

// Nudge returns a [Nudge] solution based on the lambda nudge amount
// applied according to this Jacobian.
func (j Jacobian) Nudge(lambda float64) Nudge {
	return Nudge{
		Linear:  dprec.Vec3Prod(j.LinearSlope, lambda),
		Angular: dprec.Vec3Prod(j.AngularSlope, lambda),
	}
}
