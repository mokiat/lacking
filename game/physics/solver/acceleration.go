package solver

import (
	"github.com/mokiat/gomath/dprec"
)

type Acceleration interface {
	ApplyAcceleration(ctx AccelerationContext)
}

type AccelerationContext struct {
	Target *AccelerationTarget
}

func NewAccelerationTarget(linearVelocity, angularVelocity dprec.Vec3) AccelerationTarget {
	return AccelerationTarget{
		linearVelocity:  linearVelocity,
		angularVelocity: angularVelocity,
	}
}

type AccelerationTarget struct {
	linearVelocity  dprec.Vec3
	angularVelocity dprec.Vec3

	linearAcceleration  dprec.Vec3
	angularAcceleration dprec.Vec3
}

func (t *AccelerationTarget) LinearVelocity() dprec.Vec3 {
	return t.linearVelocity
}

func (t *AccelerationTarget) AngularVelocity() dprec.Vec3 {
	return t.angularVelocity
}

func (t *AccelerationTarget) AddLinearAcceleration(acceleration dprec.Vec3) {
	t.linearAcceleration = dprec.Vec3Sum(t.linearAcceleration, acceleration)
}

func (t *AccelerationTarget) AccumulatedLinearAcceleration() dprec.Vec3 {
	return t.linearAcceleration
}

func (t *AccelerationTarget) AddAngularAcceleration(acceleration dprec.Vec3) {
	t.angularAcceleration = dprec.Vec3Sum(t.angularAcceleration, acceleration)
}

func (t *AccelerationTarget) AccumulatedAngularAcceleration() dprec.Vec3 {
	return t.angularAcceleration
}
