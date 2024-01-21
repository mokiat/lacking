package physics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/util/spatial"
)

type BodyDefinitionInfo struct {
	Mass                   float64
	MomentOfInertia        dprec.Mat3
	FrictionCoefficient    float64
	RestitutionCoefficient float64
	DragFactor             float64
	AngularDragFactor      float64
	CollisionGroup         int
	CollisionSpheres       []collision.Sphere
	CollisionBoxes         []collision.Box
	CollisionMeshes        []collision.Mesh
	AerodynamicShapes      []AerodynamicShape
}

type BodyDefinition struct {
	mass                   float64
	momentOfInertia        dprec.Mat3
	frictionCoefficient    float64
	restitutionCoefficient float64
	dragFactor             float64
	angularDragFactor      float64
	collisionGroup         int
	collisionSet           collision.Set
	aerodynamicShapes      []AerodynamicShape
}

func (d *BodyDefinition) CollisionSet() collision.Set {
	return d.collisionSet
}

type BodyInfo struct {
	Name       string
	Definition *BodyDefinition
	Position   dprec.Vec3
	Rotation   dprec.Quat
}

var freeBodyID uint32

// Body represents a physical body that has physics
// act upon it.
type Body struct {
	id     uint32
	scene  *Scene
	itemID spatial.DynamicOctreeItemID

	revision   int
	definition *BodyDefinition

	name string

	mass                   float64
	momentOfInertia        dprec.Mat3
	frictionCoefficient    float64
	restitutionCoefficient float64
	dragFactor             float64
	angularDragFactor      float64

	oldPosition    dprec.Vec3
	oldOrientation dprec.Quat

	position    dprec.Vec3
	orientation dprec.Quat

	lerpPosition     dprec.Vec3
	slerpOrientation dprec.Quat

	acceleration        dprec.Vec3
	angularAcceleration dprec.Vec3

	velocity        dprec.Vec3
	angularVelocity dprec.Vec3

	bsRadius          float64
	collisionGroup    int
	collisionSet      collision.Set
	aerodynamicShapes []AerodynamicShape
}

func (b *Body) invalidateCollisionShapes() {
	transform := collision.TRTransform(b.position, b.orientation)
	b.collisionSet.Replace(b.definition.collisionSet, transform)

	bs := b.collisionSet.BoundingSphere()
	delta := dprec.Vec3Diff(bs.Position(), b.position)
	b.bsRadius = delta.Length() + bs.Radius()
	b.scene.bodyOctree.Update(b.itemID, b.position, b.bsRadius)
}

// Name returns the name of this body.
func (b *Body) Name() string {
	return b.name
}

// SetName sets a new name for this body.
func (b *Body) SetName(name string) {
	b.name = name
}

// Mass returns the mass of this body in kg.
func (b *Body) Mass() float64 {
	return b.mass
}

// SetMass changes the mass of this body.
func (b *Body) SetMass(mass float64) {
	b.mass = mass
}

// MomentOfInertia returns the moment of inertia, or
// rotational inertia of this body.
func (b *Body) MomentOfInertia() dprec.Mat3 {
	return b.momentOfInertia
}

// SetMomentOfInertia changes the moment of inertia
// of this body.
func (b *Body) SetMomentOfInertia(inertia dprec.Mat3) {
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
func (b *Body) RestitutionCoefficient() float64 {
	return b.restitutionCoefficient
}

// SetRestitutionCoefficient changes the restitution
// coefficient for this body.
func (b *Body) SetRestitutionCoefficient(coefficient float64) {
	b.restitutionCoefficient = coefficient
}

// DragCoefficient returns the drag factor of this body.
func (b *Body) DragFactor() float64 {
	return b.dragFactor
}

// SetDragFactor sets the drag factor for this body.
// The drag factor is the drag coefficient multiplied
// by the area and divided in half.
func (b *Body) SetDragFactor(factor float64) {
	b.dragFactor = factor
}

// AngularDragFactor returns the angular drag factor
// for this body.
func (b *Body) AngularDragFactor() float64 {
	return b.angularDragFactor
}

// SetAngularDragFactor sets the angular factor for this body.
// The angular factor is similar to the drag factor, except
// that it deals with the drag induced by the rotation of
// the body.
func (b *Body) SetAngularDragFactor(factor float64) {
	b.angularDragFactor = factor
}

// Position returns the body's position in world
// space.
func (b *Body) Position() dprec.Vec3 {
	return b.position
}

// SetPosition changes the position of this body.
func (b *Body) SetPosition(position dprec.Vec3) {
	b.position = position
	b.lerpPosition = position
	b.scene.bodyOctree.Update(b.itemID, b.position, b.bsRadius)
	// TODO: Do this only on demand.
	b.invalidateCollisionShapes()
}

// VisualPosition returns the position of the Body as would
// be seen by the current frame.
//
// NOTE: The physics engine can advance past the current frame,
// which is the reason for this method.
func (b *Body) VisualPosition() dprec.Vec3 {
	return b.lerpPosition
}

// Orientation returns the quaternion rotation
// of this body.
func (b *Body) Orientation() dprec.Quat {
	return b.orientation
}

// SetOrientation changes the quaterntion rotation
// of this body.
func (b *Body) SetOrientation(orientation dprec.Quat) {
	b.orientation = orientation
	b.slerpOrientation = orientation
}

// VisualOrientation returns the orientation of the Body as would
// be seen by the current frame.
//
// NOTE: The physics engine can advance past the current frame,
// which is the reason for this method.
func (b *Body) VisualOrientation() dprec.Quat {
	return b.slerpOrientation
}

// Velocity returns the velocity of this body.
func (b *Body) Velocity() dprec.Vec3 {
	return b.velocity
}

// SetVelocity changes the velocity of this body.
func (b *Body) SetVelocity(velocity dprec.Vec3) {
	b.velocity = velocity
}

// AngularVelocity returns the angular velocity
// of this body.
func (b *Body) AngularVelocity() dprec.Vec3 {
	return b.angularVelocity
}

// SetAngularVelocity changes the angular velocity
// of this body.
func (b *Body) SetAngularVelocity(angularVelocity dprec.Vec3) {
	b.angularVelocity = angularVelocity
}

// CollisionGroup returns the collision group for this body. Two bodies
// with the same collision group are not checked for collisions.
func (b *Body) CollisionGroup() int {
	return b.collisionGroup
}

// SetCollisionGroup changes the collision group for this body.
//
// A value of 0 disables the collision group.
func (b *Body) SetCollisionGroup(group int) {
	b.collisionGroup = group
}

// CollisionSet contains the collision shapes for this body.
func (b *Body) CollisionSet() collision.Set {
	return b.collisionSet
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
	delete(b.scene.dynamicBodies, b)
	b.scene.bodyOctree.Remove(b.itemID)
	b.scene.bodyPool.Restore(b)
	b.collisionSet = collision.Set{}
	b.scene = nil
}

func (b *Body) resetAcceleration() {
	b.acceleration = dprec.ZeroVec3()
}

func (b *Body) clampAcceleration(max float64) {
	if b.acceleration.SqrLength() > max*max {
		b.acceleration = dprec.ResizedVec3(b.acceleration, max)
	}
}

func (b *Body) resetAngularAcceleration() {
	b.angularAcceleration = dprec.ZeroVec3()
}

func (b *Body) clampAngularAcceleration(max float64) {
	if b.angularAcceleration.SqrLength() > max*max {
		b.angularAcceleration = dprec.ResizedVec3(b.angularAcceleration, max)
	}
}

func (b *Body) addAcceleration(amount dprec.Vec3) {
	b.acceleration = dprec.Vec3Sum(b.acceleration, amount)
}

func (b *Body) addAngularAcceleration(amount dprec.Vec3) {
	b.angularAcceleration = dprec.Vec3Sum(b.angularAcceleration, amount)
}

func (b *Body) applyForce(force dprec.Vec3) {
	b.addAcceleration(dprec.Vec3Quot(force, b.mass))
}

func (b *Body) applyOffsetForce(offset, force dprec.Vec3) {
	b.applyForce(force)
	b.applyTorque(dprec.Vec3Cross(offset, force))
}

func (b *Body) applyTorque(torque dprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the torque is in world space
	b.addAngularAcceleration(dprec.Mat3Vec3Prod(dprec.InverseMat3(b.momentOfInertia), torque))
}

func (b *Body) clampVelocity(max float64) {
	if b.velocity.SqrLength() > max*max {
		b.velocity = dprec.ResizedVec3(b.velocity, max)
	}
}

func (b *Body) clampAngularVelocity(max float64) {
	if b.angularVelocity.SqrLength() > max*max {
		b.angularVelocity = dprec.ResizedVec3(b.angularVelocity, max)
	}
}

func (b *Body) addVelocity(amount dprec.Vec3) {
	b.velocity = dprec.Vec3Sum(b.velocity, amount)
}

func (b *Body) addAngularVelocity(amount dprec.Vec3) {
	b.angularVelocity = dprec.Vec3Sum(b.angularVelocity, amount)
}

func (b *Body) applyImpulse(impulse dprec.Vec3) {
	b.addVelocity(dprec.Vec3Quot(impulse, b.mass))
}

func (b *Body) applyAngularImpulse(impulse dprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the impulse is in world space
	b.addAngularVelocity(dprec.Mat3Vec3Prod(dprec.InverseMat3(b.momentOfInertia), impulse))
}

func (b *Body) applyOffsetImpulse(offset, impulse dprec.Vec3) {
	b.applyImpulse(impulse)
	b.applyAngularImpulse(dprec.Vec3Cross(offset, impulse))
}

func (b *Body) translate(offset dprec.Vec3) {
	b.SetPosition(dprec.Vec3Sum(b.position, offset))
}

func (b *Body) rotate(quat dprec.Quat) {
	b.orientation = dprec.UnitQuat(dprec.QuatProd(quat, b.orientation))
}

func (b *Body) vectorRotate(vector dprec.Vec3) {
	const angularEpsilon = float64(0.00001)
	if radians := vector.Length(); dprec.Abs(radians) > angularEpsilon {
		b.rotate(dprec.RotationQuat(dprec.Radians(radians), vector))
	}
}

func (b *Body) applyNudge(nudge dprec.Vec3) {
	b.translate(dprec.Vec3Quot(nudge, b.mass))
}

func (b *Body) applyAngularNudge(nudge dprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the torque is in world space
	b.vectorRotate(dprec.Mat3Vec3Prod(dprec.InverseMat3(b.momentOfInertia), nudge))
}

func (b *Body) applyOffsetNudge(offset, nudge dprec.Vec3) {
	b.applyNudge(nudge)
	b.applyAngularNudge(dprec.Vec3Cross(offset, nudge))
}
