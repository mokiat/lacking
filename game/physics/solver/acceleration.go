package solver

import "github.com/mokiat/gomath/dprec"

// Acceleration represents a solver that can apply acceleration
// to a target.
type Acceleration interface {

	// ApplyAcceleration applies acceleration to the target.
	ApplyAcceleration(ctx AccelerationContext)
}

// AccelerationContext provides information about the target
// that is being accelerated.
type AccelerationContext struct {
	Target *AccelerationTarget
}

// NewAccelerationTarget creates a new AccelerationTarget.
func NewAccelerationTarget(
	mass float64,
	momentOfInertia dprec.Mat3,
	position dprec.Vec3,
	rotation dprec.Quat,
	linearVelocity dprec.Vec3,
	angularVelocity dprec.Vec3,
) AccelerationTarget {
	return AccelerationTarget{
		mass:            mass,
		momentOfInertia: momentOfInertia,
		position:        position,
		rotation:        rotation,
		linearVelocity:  linearVelocity,
		angularVelocity: angularVelocity,
	}
}

// AccelerationTarget represents a target that can be accelerated.
type AccelerationTarget struct {
	mass            float64
	momentOfInertia dprec.Mat3

	position dprec.Vec3
	rotation dprec.Quat

	linearVelocity  dprec.Vec3
	angularVelocity dprec.Vec3

	linearAcceleration  dprec.Vec3
	angularAcceleration dprec.Vec3
}

// Mass returns the mass of the target.
func (t *AccelerationTarget) Mass() float64 {
	return t.mass
}

// Position returns the position of the target.
func (t *AccelerationTarget) Position() dprec.Vec3 {
	return t.position
}

// Rotation returns the rotation of the target.
func (t *AccelerationTarget) Rotation() dprec.Quat {
	return t.rotation
}

// LinearVelocity returns the linear velocity of the target.
func (t *AccelerationTarget) LinearVelocity() dprec.Vec3 {
	return t.linearVelocity
}

// AngularVelocity returns the angular velocity of the target.
func (t *AccelerationTarget) AngularVelocity() dprec.Vec3 {
	return t.angularVelocity
}

// AccumulatedLinearAcceleration returns the accumulated linear acceleration
// of the target.
func (t *AccelerationTarget) AccumulatedLinearAcceleration() dprec.Vec3 {
	return t.linearAcceleration
}

// ApplyLinearAcceleration adds linear acceleration to the target.
func (t *AccelerationTarget) ApplyLinearAcceleration(acceleration dprec.Vec3) {
	t.linearAcceleration = dprec.Vec3Sum(t.linearAcceleration, acceleration)
}

// TODO: Apply force

// AccumulatedAngularAcceleration returns the accumulated angular acceleration
// of the target.
func (t *AccelerationTarget) AccumulatedAngularAcceleration() dprec.Vec3 {
	return t.angularAcceleration
}

// ApplyAngularAcceleration adds angular acceleration to the target.
func (t *AccelerationTarget) ApplyAngularAcceleration(acceleration dprec.Vec3) {
	t.angularAcceleration = dprec.Vec3Sum(t.angularAcceleration, acceleration)
}

// TODO: Apply torque

// TODO: Apply offset force
