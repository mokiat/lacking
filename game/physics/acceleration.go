package physics

import "github.com/mokiat/gomath/dprec"

// AccelerationTarget represents an object that can be accelerated.
//
// It exposes the read-only motion state of the object and accumulates the
// linear and angular acceleration that is to be applied to it. The
// accumulated acceleration does not affect the state that is observed
// through the target, which means that the order in which independent
// contributors are evaluated does not matter.
//
// All vectors and tensors are expressed in world space, with offsets being
// relative to the object's center of mass.
type AccelerationTarget struct {
	body *bodyState
}

// newAccelerationTarget creates a new [AccelerationTarget] backed by the
// specified body state.
func newAccelerationTarget(body *bodyState) AccelerationTarget {
	return AccelerationTarget{
		body: body,
	}
}

// InverseMass returns the reciprocal of the mass of the target. A value of
// zero indicates an immovable object of infinite mass.
//
// Prefer this method over [AccelerationTarget.Mass], since the physics
// engine works with the inverse mass internally and no division is needed.
func (t AccelerationTarget) InverseMass() float64 {
	return t.body.invMass
}

// Mass returns the mass of the target.
//
// This method performs a division and returns positive infinity for an
// immovable object. Prefer [AccelerationTarget.InverseMass] where possible.
func (t AccelerationTarget) Mass() float64 {
	return t.body.mass()
}

// InverseInertia returns the reciprocal of the moment of inertia tensor of
// the target in world space. A zero matrix indicates an object that cannot
// be rotated.
//
// Prefer this method over [AccelerationTarget.Inertia], since the physics
// engine works with the inverse tensor internally and no matrix inversion
// is needed.
func (t AccelerationTarget) InverseInertia() dprec.Mat3 {
	return t.body.invInertia
}

// Inertia returns the moment of inertia tensor of the target in world space.
//
// This method performs a matrix inversion and is undefined for an object
// that cannot be rotated. Prefer [AccelerationTarget.InverseInertia] where
// possible.
func (t AccelerationTarget) Inertia() dprec.Mat3 {
	return t.body.inertia()
}

// Position returns the world position of the center of mass of the target.
func (t AccelerationTarget) Position() dprec.Vec3 {
	return t.body.position
}

// Rotation returns the world orientation of the target.
func (t AccelerationTarget) Rotation() dprec.Quat {
	return t.body.rotation
}

// LinearVelocity returns the world velocity of the center of mass of the
// target.
func (t AccelerationTarget) LinearVelocity() dprec.Vec3 {
	return t.body.linearVelocity
}

// AngularVelocity returns the world angular velocity of the target, in
// radians per second around each axis.
func (t AccelerationTarget) AngularVelocity() dprec.Vec3 {
	return t.body.angularVelocity
}

// LinearAcceleration returns the linear acceleration that has been
// accumulated on the target so far.
func (t AccelerationTarget) LinearAcceleration() dprec.Vec3 {
	return t.body.linearAcceleration
}

// AddLinearAcceleration accumulates the specified linear acceleration on the
// target.
//
// Since acceleration is independent of mass, this is the correct way to
// model effects like gravity, which pull on all objects equally.
func (t AccelerationTarget) AddLinearAcceleration(acceleration dprec.Vec3) {
	t.body.addLinearAcceleration(acceleration)
}

// ApplyForce accumulates the linear acceleration that results from the
// specified force acting on the center of mass of the target.
//
// Use [AccelerationTarget.ApplyOffsetForce] instead if the force does not
// act on the center of mass and should induce rotation.
func (t AccelerationTarget) ApplyForce(force dprec.Vec3) {
	t.body.applyForce(force)
}

// AngularAcceleration returns the angular acceleration that has been
// accumulated on the target so far.
func (t AccelerationTarget) AngularAcceleration() dprec.Vec3 {
	return t.body.angularAcceleration
}

// AddAngularAcceleration accumulates the specified angular acceleration on
// the target.
func (t AccelerationTarget) AddAngularAcceleration(acceleration dprec.Vec3) {
	t.body.addAngularAcceleration(acceleration)
}

// ApplyTorque accumulates the angular acceleration that results from the
// specified torque acting on the target.
func (t AccelerationTarget) ApplyTorque(torque dprec.Vec3) {
	t.body.applyTorque(torque)
}

// ApplyOffsetForce accumulates the linear and angular acceleration that
// result from the specified force acting on the target at the specified
// offset from its center of mass.
func (t AccelerationTarget) ApplyOffsetForce(offset, force dprec.Vec3) {
	t.body.applyOffsetForce(offset, force)
}

// AccelerationContext carries everything an [AccelerationSolver] needs in
// order to accelerate a single target during a physics simulation step.
//
// Alongside the target itself, it carries the parts of the surrounding
// medium at the target's location that are shared by every contributor
// evaluated against that target this step, so that each of them does not
// have to sample the medium on its own.
type AccelerationContext struct {

	// DeltaSeconds is the time step of the simulation, in seconds. Use this
	// in case of any time-dependent effects, like drag.
	DeltaSeconds float64

	// MediumVelocity is the velocity of the medium in world space, in m/s.
	MediumVelocity dprec.Vec3

	// MediumDensity is the density of the medium in kg/m^3.
	MediumDensity float64

	// Target is the [AccelerationTarget] being accelerated, through which
	// the solver reads its motion state and accumulates the acceleration
	// it contributes.
	Target AccelerationTarget
}

// AccelerationSolver applies an acceleration effect to a target.
//
// Implementations model a single physical effect, like gravity or thrust,
// and are evaluated once per body per simulation step.
type AccelerationSolver interface {

	// ApplyAcceleration accumulates on ctx.Target the acceleration that this
	// effect produces on it, under the rest of ctx.
	//
	// Implementations must not retain ctx or ctx.Target, since both are only
	// valid for the duration of the call, and must not mutate any state that
	// other contributors observe, since the evaluation order is unspecified.
	ApplyAcceleration(ctx AccelerationContext)
}
