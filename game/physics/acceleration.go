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
	invMass    float64
	invInertia dprec.Mat3

	position dprec.Vec3
	rotation dprec.Quat

	linearVelocity  dprec.Vec3
	angularVelocity dprec.Vec3

	linearAcceleration  dprec.Vec3
	angularAcceleration dprec.Vec3
}

// newAccelerationTarget creates a new [AccelerationTarget] with the specified
// state.
//
// The invMass and invInertia parameters are the reciprocals of the mass and
// of the moment of inertia tensor respectively. A value of zero (or a zero
// matrix) represents an immovable object of infinite mass or inertia.
//
// The inverse moment of inertia has to be expressed in world space. Use
// [RotatedMomentOfInertia] with the object's rotation to bring a tensor that
// is expressed in local space into world space before inverting it.
func newAccelerationTarget(
	invMass float64,
	invInertia dprec.Mat3,
	position dprec.Vec3,
	rotation dprec.Quat,
	linearVelocity dprec.Vec3,
	angularVelocity dprec.Vec3,
) AccelerationTarget {
	return AccelerationTarget{
		invMass:         invMass,
		invInertia:      invInertia,
		position:        position,
		rotation:        rotation,
		linearVelocity:  linearVelocity,
		angularVelocity: angularVelocity,
	}
}

// InverseMass returns the reciprocal of the mass of the target. A value of
// zero indicates an immovable object of infinite mass.
//
// Prefer this method over [AccelerationTarget.Mass], since the physics
// engine works with the inverse mass internally and no division is needed.
func (t *AccelerationTarget) InverseMass() float64 {
	return t.invMass
}

// Mass returns the mass of the target.
//
// This method performs a division and returns positive infinity for an
// immovable object. Prefer [AccelerationTarget.InverseMass] where possible.
func (t *AccelerationTarget) Mass() float64 {
	return 1.0 / t.invMass
}

// InverseInertia returns the reciprocal of the moment of inertia tensor of
// the target in world space. A zero matrix indicates an object that cannot
// be rotated.
//
// Prefer this method over [AccelerationTarget.Inertia], since the physics
// engine works with the inverse tensor internally and no matrix inversion
// is needed.
func (t *AccelerationTarget) InverseInertia() dprec.Mat3 {
	return t.invInertia
}

// Inertia returns the moment of inertia tensor of the target in world space.
//
// This method performs a matrix inversion and is undefined for an object
// that cannot be rotated. Prefer [AccelerationTarget.InverseInertia] where
// possible.
func (t *AccelerationTarget) Inertia() dprec.Mat3 {
	return dprec.InverseMat3(t.invInertia)
}

// Position returns the world position of the center of mass of the target.
func (t *AccelerationTarget) Position() dprec.Vec3 {
	return t.position
}

// Rotation returns the world orientation of the target.
func (t *AccelerationTarget) Rotation() dprec.Quat {
	return t.rotation
}

// LinearVelocity returns the world velocity of the center of mass of the
// target.
func (t *AccelerationTarget) LinearVelocity() dprec.Vec3 {
	return t.linearVelocity
}

// AngularVelocity returns the world angular velocity of the target, in
// radians per second around each axis.
func (t *AccelerationTarget) AngularVelocity() dprec.Vec3 {
	return t.angularVelocity
}

// LinearAcceleration returns the linear acceleration that has been
// accumulated on the target so far.
func (t *AccelerationTarget) LinearAcceleration() dprec.Vec3 {
	return t.linearAcceleration
}

// AddLinearAcceleration accumulates the specified linear acceleration on the
// target.
//
// Since acceleration is independent of mass, this is the correct way to
// model effects like gravity, which pull on all objects equally.
func (t *AccelerationTarget) AddLinearAcceleration(acceleration dprec.Vec3) {
	t.linearAcceleration = dprec.Vec3Sum(t.linearAcceleration, acceleration)
}

// ApplyForce accumulates the linear acceleration that results from the
// specified force acting on the center of mass of the target.
//
// Use [AccelerationTarget.ApplyOffsetForce] instead if the force does not
// act on the center of mass and should induce rotation.
func (t *AccelerationTarget) ApplyForce(force dprec.Vec3) {
	t.AddLinearAcceleration(dprec.Vec3Prod(force, t.invMass))
}

// AngularAcceleration returns the angular acceleration that has been
// accumulated on the target so far.
func (t *AccelerationTarget) AngularAcceleration() dprec.Vec3 {
	return t.angularAcceleration
}

// AddAngularAcceleration accumulates the specified angular acceleration on
// the target.
func (t *AccelerationTarget) AddAngularAcceleration(acceleration dprec.Vec3) {
	t.angularAcceleration = dprec.Vec3Sum(t.angularAcceleration, acceleration)
}

// ApplyTorque accumulates the angular acceleration that results from the
// specified torque acting on the target.
func (t *AccelerationTarget) ApplyTorque(torque dprec.Vec3) {
	t.AddAngularAcceleration(dprec.Mat3Vec3Prod(t.invInertia, torque))
}

// ApplyOffsetForce accumulates the linear and angular acceleration that
// result from the specified force acting on the target at the specified
// offset from its center of mass.
func (t *AccelerationTarget) ApplyOffsetForce(offset, force dprec.Vec3) {
	t.ApplyForce(force)
	t.ApplyTorque(dprec.Vec3Cross(offset, force))
}

// AccelerationContext describes the surrounding medium at the location of
// the target that is being accelerated.
//
// It carries the parts of the environment that are shared by all
// contributors, so that each of them does not have to sample the medium on
// its own.
type AccelerationContext struct {

	// MediumVelocity is the velocity of the medium in world space, in m/s.
	MediumVelocity dprec.Vec3

	// MediumDensity is the density of the medium in kg/m^3.
	MediumDensity float64
}

// AccelerationSolver applies an acceleration effect to a target.
//
// Implementations model a single physical effect, like gravity or thrust,
// and are evaluated once per body per simulation step.
type AccelerationSolver interface {

	// ApplyAcceleration accumulates on the target the acceleration that this
	// effect produces on it under the specified context.
	//
	// Implementations must not retain the target, since it is only valid for
	// the duration of the call, and must not mutate any state that other
	// contributors observe, since the evaluation order is unspecified.
	ApplyAcceleration(ctx AccelerationContext, target *AccelerationTarget)
}
