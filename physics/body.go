package physics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/shape"
)

type Body struct {
	Name string

	Position    sprec.Vec3
	Orientation sprec.Quat
	IsStatic    bool

	Mass            float32
	MomentOfInertia sprec.Mat3

	Acceleration        sprec.Vec3
	AngularAcceleration sprec.Vec3

	Velocity        sprec.Vec3
	AngularVelocity sprec.Vec3

	DragFactor        float32
	AngularDragFactor float32

	RestitutionCoef float32
	CollisionShapes []shape.Placement
	InCollision     bool
}

func (b *Body) Translate(offset sprec.Vec3) {
	b.Position = sprec.Vec3Sum(b.Position, offset)
}

func (b *Body) Rotate(vector sprec.Vec3) {
	if radians := vector.Length(); sprec.Abs(radians) > radialEpsilon {
		b.Orientation = sprec.QuatProd(sprec.RotationQuat(sprec.Radians(radians), vector), b.Orientation)
	}
}

func (b *Body) ApplyNudge(nudge sprec.Vec3) {
	b.Translate(sprec.Vec3Quot(nudge, b.Mass))
}

func (b *Body) ApplyAngularNudge(nudge sprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the torque is in world space
	b.Rotate(sprec.Mat3Vec3Prod(sprec.InverseMat3(b.MomentOfInertia), nudge))
}

func (b *Body) ResetAcceleration() {
	b.Acceleration = sprec.ZeroVec3()
}

func (b *Body) ResetAngularAcceleration() {
	b.AngularAcceleration = sprec.ZeroVec3()
}

func (b *Body) AddAcceleration(amount sprec.Vec3) {
	b.Acceleration = sprec.Vec3Sum(b.Acceleration, amount)
}

func (b *Body) AddAngularAcceleration(amount sprec.Vec3) {
	b.AngularAcceleration = sprec.Vec3Sum(b.AngularAcceleration, amount)
}

func (b *Body) AddVelocity(amount sprec.Vec3) {
	b.Velocity = sprec.Vec3Sum(b.Velocity, amount)
}

func (b *Body) AddAngularVelocity(amount sprec.Vec3) {
	b.AngularVelocity = sprec.Vec3Sum(b.AngularVelocity, amount)
}

func (b *Body) ApplyForce(force sprec.Vec3) {
	b.AddAcceleration(sprec.Vec3Quot(force, b.Mass))
}

func (b *Body) ApplyTorque(torque sprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the torque is in world space
	b.AddAngularAcceleration(sprec.Mat3Vec3Prod(sprec.InverseMat3(b.MomentOfInertia), torque))
}

func (b *Body) ApplyOffsetForce(offset, force sprec.Vec3) {
	b.ApplyForce(force)
	b.ApplyTorque(sprec.Vec3Cross(offset, force))
}

func (b *Body) ApplyImpulse(impulse sprec.Vec3) {
	b.Velocity = sprec.Vec3Sum(b.Velocity, sprec.Vec3Quot(impulse, b.Mass))
}

func (b *Body) ApplyAngularImpulse(impulse sprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the impulse is in world space
	b.AddAngularVelocity(sprec.Mat3Vec3Prod(sprec.InverseMat3(b.MomentOfInertia), impulse))
}

func (b *Body) ApplyOffsetImpulse(offset, impulse sprec.Vec3) {
	b.ApplyImpulse(impulse)
	b.ApplyAngularImpulse(sprec.Vec3Cross(offset, impulse))
}

func SymmetricMomentOfInertia(value float32) sprec.Mat3 {
	return sprec.NewMat3(
		value, 0.0, 0.0,
		0.0, value, 0.0,
		0.0, 0.0, value,
	)
}
