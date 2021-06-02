package physics

import "github.com/mokiat/gomath/sprec"

// Body represents a physical body that has physics
// act upon it.
type Body struct {
	scene *Scene
	prev  *Body
	next  *Body

	name string

	static                 bool
	mass                   float32
	momentOfInertia        sprec.Mat3
	restitutionCoefficient float32
	dragFactor             float32
	angularDragFactor      float32

	position    sprec.Vec3
	orientation sprec.Quat

	acceleration        sprec.Vec3
	angularAcceleration sprec.Vec3

	velocity        sprec.Vec3
	angularVelocity sprec.Vec3

	collisionShapes   []CollisionShape
	aerodynamicShapes []AerodynamicShape
}

// Name returns the name of this body.
func (b *Body) Name() string {
	return b.name
}

// SetName sets a new name for this body.
func (b *Body) SetName(name string) {
	b.name = name
}

// Static returns whether this body is static.
func (b *Body) Static() bool {
	return b.static
}

// SetStatic changes whether this body is static.
func (b *Body) SetStatic(static bool) {
	b.static = static
}

// Mass returns the mass of this body in kg.
func (b *Body) Mass() float32 {
	return b.mass
}

// SetMass changes the mass of this body.
func (b *Body) SetMass(mass float32) {
	b.mass = mass
}

// MomentOfInertia returns the moment of inertia, or
// rotational inertia of this body.
func (b *Body) MomentOfInertia() sprec.Mat3 {
	return b.momentOfInertia
}

// SetMomentOfInertia changes the moment of inertia
// of this body.
func (b *Body) SetMomentOfInertia(inertia sprec.Mat3) {
	b.momentOfInertia = inertia
}

// RestitutionCoefficient returns the restitution
// coefficient of this body. Valid values are in
// the range [0.0 - 1.0], where 0.0 means that the
// body does not bounce and 1.0 means that it bounds
// back with the same velocity. In reality the amount
// that the body will bounce depends on the restitution
// coefficients of both bodies colliding. Furthermore,
// due to computational errors, the bounce will eventually
// stop.
func (b *Body) RestitutionCoefficient() float32 {
	return b.restitutionCoefficient
}

// SetRestitutionCoefficient changes the restitution
// coefficient for this body.
func (b *Body) SetRestitutionCoefficient(coefficient float32) {
	b.restitutionCoefficient = coefficient
}

// DragCoefficient returns the drag factor of this body.
func (b *Body) DragFactor() float32 {
	return b.dragFactor
}

// SetDragFactor sets the drag factor for this body.
// The drag factor is the drag coefficient multiplied
// by the area and divided in half.
func (b *Body) SetDragFactor(factor float32) {
	b.dragFactor = factor
}

// AngularDragFactor returns the angular drag factor
// for this body.
func (b *Body) AngularDragFactor() float32 {
	return b.angularDragFactor
}

// SetAngularDragFactor sets the angular factor for this body.
// The angular factor is similar to the drag factor, except
// that it deals with the drag induced by the rotation of
// the body.
func (b *Body) SetAngularDragFactor(factor float32) {
	b.angularDragFactor = factor
}

// Position returns the body's position in world
// space.
func (b *Body) Position() sprec.Vec3 {
	return b.position
}

// SetPosition changes the position of this body.
func (b *Body) SetPosition(position sprec.Vec3) {
	b.position = position
}

// Orientation returns the quaternion rotation
// of this body.
func (b *Body) Orientation() sprec.Quat {
	return b.orientation
}

// SetOrientation changes the quaterntion rotation
// of this body.
func (b *Body) SetOrientation(orientation sprec.Quat) {
	b.orientation = orientation
}

// Velocity returns the velocity of this body.
func (b *Body) Velocity() sprec.Vec3 {
	return b.velocity
}

// SetVelocity changes the velocity of this body.
func (b *Body) SetVelocity(velocity sprec.Vec3) {
	b.velocity = velocity
}

// AngularVelocity returns the angular velocity
// of this body.
func (b *Body) AngularVelocity() sprec.Vec3 {
	return b.angularVelocity
}

// SetAngularVelocity changes the angular velocity
// of this body.
func (b *Body) SetAngularVelocity(angularVelocity sprec.Vec3) {
	b.angularVelocity = angularVelocity
}

// CollisionShapes returns a slice of shapes that
// dictate how this body collides with others.
func (b *Body) CollisionShapes() []CollisionShape {
	return b.collisionShapes
}

// SetCollisionShapes sets the collision shapes
// for this body to be used in collision detection.
func (b *Body) SetCollisionShapes(shapes []CollisionShape) {
	b.collisionShapes = shapes
}

// AerodynamicShapes returns a slice of shapes that
// dictate how this body is affected by relative air
// motion.
func (b *Body) AerodynamicShapes() []AerodynamicShape {
	return b.aerodynamicShapes
}

// SetAerodynamicShapes sets the aerodynamics shapes
// to be used when calculating wind drag and lift.
func (b *Body) SetAerodynamicShapes(shapes []AerodynamicShape) {
	b.aerodynamicShapes = shapes
}

// Delete removes this physical body. The object
// should no longer be used after calling this
// method.
func (b *Body) Delete() {
	b.scene.removeBody(b)
	b.scene.cacheBody(b)
	b.scene = nil
}

func (b *Body) resetAcceleration() {
	b.acceleration = sprec.ZeroVec3()
}

func (b *Body) clampAcceleration(max float32) {
	if b.acceleration.SqrLength() > max*max {
		b.acceleration = sprec.ResizedVec3(b.acceleration, max)
	}
}

func (b *Body) resetAngularAcceleration() {
	b.angularAcceleration = sprec.ZeroVec3()
}

func (b *Body) clampAngularAcceleration(max float32) {
	if b.angularAcceleration.SqrLength() > max*max {
		b.angularAcceleration = sprec.ResizedVec3(b.angularAcceleration, max)
	}
}

func (b *Body) addAcceleration(amount sprec.Vec3) {
	b.acceleration = sprec.Vec3Sum(b.acceleration, amount)
}

func (b *Body) addAngularAcceleration(amount sprec.Vec3) {
	b.angularAcceleration = sprec.Vec3Sum(b.angularAcceleration, amount)
}

func (b *Body) applyForce(force sprec.Vec3) {
	b.addAcceleration(sprec.Vec3Quot(force, b.mass))
}

func (b *Body) applyTorque(torque sprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the torque is in world space
	b.addAngularAcceleration(sprec.Mat3Vec3Prod(sprec.InverseMat3(b.momentOfInertia), torque))
}

func (b *Body) clampVelocity(max float32) {
	if b.velocity.SqrLength() > max*max {
		b.velocity = sprec.ResizedVec3(b.velocity, max)
	}
}

func (b *Body) clampAngularVelocity(max float32) {
	if b.angularVelocity.SqrLength() > max*max {
		b.angularVelocity = sprec.ResizedVec3(b.angularVelocity, max)
	}
}

func (b *Body) addVelocity(amount sprec.Vec3) {
	b.velocity = sprec.Vec3Sum(b.velocity, amount)
}

func (b *Body) addAngularVelocity(amount sprec.Vec3) {
	b.angularVelocity = sprec.Vec3Sum(b.angularVelocity, amount)
}

func (b *Body) applyImpulse(impulse sprec.Vec3) {
	b.addVelocity(sprec.Vec3Quot(impulse, b.mass))
}

func (b *Body) applyAngularImpulse(impulse sprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the impulse is in world space
	b.addAngularVelocity(sprec.Mat3Vec3Prod(sprec.InverseMat3(b.momentOfInertia), impulse))
}

func (b *Body) applyOffsetImpulse(offset, impulse sprec.Vec3) {
	b.applyImpulse(impulse)
	b.applyAngularImpulse(sprec.Vec3Cross(offset, impulse))
}

func (b *Body) translate(offset sprec.Vec3) {
	b.position = sprec.Vec3Sum(b.position, offset)
}

func (b *Body) rotate(quat sprec.Quat) {
	b.orientation = sprec.UnitQuat(sprec.QuatProd(quat, b.orientation))
}

func (b *Body) vectorRotate(vector sprec.Vec3) {
	const angularEpsilon = float32(0.00001)
	if radians := vector.Length(); sprec.Abs(radians) > angularEpsilon {
		b.rotate(sprec.RotationQuat(sprec.Radians(radians), vector))
	}
}

func (b *Body) applyNudge(nudge sprec.Vec3) {
	b.translate(sprec.Vec3Quot(nudge, b.mass))
}

func (b *Body) applyAngularNudge(nudge sprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the torque is in world space
	b.vectorRotate(sprec.Mat3Vec3Prod(sprec.InverseMat3(b.momentOfInertia), nudge))
}
