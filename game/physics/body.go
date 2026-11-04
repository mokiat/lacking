package physics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// BodyID uniquely identifies a body that was created through
// [BodyView.Create].
//
// The zero value, also available as [NilBodyID], does not reference a
// valid body.
type BodyID struct {
	index    int32
	revision int32
}

// NilBodyID is a [BodyID] that never references a valid body.
var NilBodyID = BodyID{}

// BodyView provides access to the bodies that belong to a [Scene].
//
// A body is a rigid object that moves under the influence of accelerators
// and constraints, and collides with other bodies and with terrains
// through the collision shapes attached to it.
//
// A BodyView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.Bodies].
type BodyView struct {
	scene *Scene
}

// Create allocates a new body at the specified world position and
// rotation, and returns its ID. The rotation must be a unit quaternion.
//
// The new body has a mass of 1, an identity moment of inertia, zero linear
// and angular velocity and no collision shapes.
func (v BodyView) Create(position dprec.Vec3, rotation dprec.Quat) BodyID {
	index, body := v.scene.allocateBody()

	objectID := v.scene.collisionScene.CreateObject(placement3d.ObjectInfo[bodyData]{
		Position: opt.V(position),
		Rotation: opt.V(rotation),
		UserData: bodyData{
			index: index,
		},
	})

	*body = bodyState{
		objectID:                  objectID,
		revision:                  body.revision + 1, // progress revision to valid (odd) value
		firstBodyAcceleratorIndex: nilIndex,
		firstSoloConstraintIndex:  nilIndex,
		firstPairConstraintIndex:  nilIndex,
		invInertiaLocal:           dprec.IdentityMat3(),
		invInertia:                dprec.IdentityMat3(),
		invMass:                   1.0,
		linearVelocity:            dprec.ZeroVec3(),
		angularVelocity:           dprec.ZeroVec3(),
		position:                  position,
		rotation:                  rotation,
	}

	return BodyID{
		index:    index,
		revision: body.revision,
	}
}

// CreateHandle behaves like [BodyView.Create] but wraps the resulting ID
// in a [BodyHandle], as returned by [BodyView.Handle], for callers that
// want to keep acting on the new body without holding onto its ID
// separately.
func (v BodyView) CreateHandle(position dprec.Vec3, rotation dprec.Quat) BodyHandle {
	return v.Handle(v.Create(position, rotation))
}

// Delete removes the body with the specified ID, together with its
// collision shapes, the body accelerators that act on it and the solo and
// pair constraints that reference it.
//
// It panics if the ID does not reference a valid body. Use
// [BodyView.IsValid] first if the ID's validity is not otherwise
// guaranteed.
func (v BodyView) Delete(id BodyID) {
	body := v.resolve(id, true)

	v.scene.collisionScene.DeleteObject(body.objectID)

	bodyAcceleratorView := v.scene.BodyAccelerators()
	for body.firstBodyAcceleratorIndex != nilIndex {
		bodyAcceleratorView.Delete(bodyAcceleratorView.idFromIndex(body.firstBodyAcceleratorIndex))
	}
	soloConstraintView := v.scene.SoloConstraints()
	for body.firstSoloConstraintIndex != nilIndex {
		soloConstraintView.Delete(soloConstraintView.idFromIndex(body.firstSoloConstraintIndex))
	}
	pairConstraintView := v.scene.PairConstraints()
	for body.firstPairConstraintIndex != nilIndex {
		pairConstraintView.Delete(pairConstraintView.idFromIndex(body.firstPairConstraintIndex))
	}

	*body = bodyState{
		objectID:                  placement3d.NilObjectID,
		revision:                  body.revision + 1, // progress revision to invalid (even) value
		firstBodyAcceleratorIndex: nilIndex,
		firstSoloConstraintIndex:  nilIndex,
		firstPairConstraintIndex:  nilIndex,
		invMass:                   1.0,
		invInertia:                dprec.IdentityMat3(),
		linearVelocity:            dprec.ZeroVec3(),
		angularVelocity:           dprec.ZeroVec3(),
		position:                  dprec.ZeroVec3(),
		rotation:                  dprec.IdentityQuat(),
	}

	v.scene.releaseBody(id.index)
}

// Each calls cb once for every body that is currently alive within this
// Scene, in unspecified order.
func (v BodyView) Each(cb func(BodyID)) {
	v.scene.eachBody(func(index int, body *bodyState) {
		cb(BodyID{
			index:    int32(index),
			revision: body.revision,
		})
	})
}

// Handle returns a [BodyHandle] that wraps the specified ID, as a more
// convenient means of repeatedly accessing the same body without having
// to pass its ID to this view on every call.
func (v BodyView) Handle(id BodyID) BodyHandle {
	return BodyHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references a body that has not
// been deleted.
func (v BodyView) IsValid(id BodyID) bool {
	body := v.resolve(id, false)
	return body != nil
}

// Mass returns the mass of the specified body, in kg.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) Mass(id BodyID) float64 {
	body := v.resolve(id, true)
	return 1.0 / body.invMass
}

// SetMass changes the mass of the specified body, in kg. The mass must be
// positive; a mass of positive infinity makes the body immovable by
// forces and impulses.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) SetMass(id BodyID, mass float64) {
	body := v.resolve(id, true)
	body.invMass = 1.0 / mass
}

// MomentOfInertia returns the moment of inertia tensor of the specified
// body, in kg*m^2, expressed in the body's local space.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) MomentOfInertia(id BodyID) dprec.Mat3 {
	body := v.resolve(id, true)
	return dprec.InverseMat3(body.invInertiaLocal)
}

// SetMomentOfInertia changes the moment of inertia tensor of the specified
// body, in kg*m^2, expressed in the body's local space. The tensor must be
// invertible. Helpers such as [SolidBoxMomentOfInertia] can be used to
// compute it.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) SetMomentOfInertia(id BodyID, inertia dprec.Mat3) {
	body := v.resolve(id, true)
	body.invInertiaLocal = dprec.InverseMat3(inertia)
	body.recalculateInertia()
}

// Velocity returns the world-space linear velocity of the center of mass
// of the specified body, in m/s.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) Velocity(id BodyID) dprec.Vec3 {
	body := v.resolve(id, true)
	return body.linearVelocity
}

// SetVelocity changes the world-space linear velocity of the center of
// mass of the specified body, in m/s.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) SetVelocity(id BodyID, velocity dprec.Vec3) {
	body := v.resolve(id, true)
	body.linearVelocity = velocity
}

// AngularVelocity returns the world-space angular velocity of the
// specified body, in radians per second around each axis.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) AngularVelocity(id BodyID) dprec.Vec3 {
	body := v.resolve(id, true)
	return body.angularVelocity
}

// SetAngularVelocity changes the world-space angular velocity of the
// specified body, in radians per second around each axis.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) SetAngularVelocity(id BodyID, angularVelocity dprec.Vec3) {
	body := v.resolve(id, true)
	body.angularVelocity = angularVelocity
}

// Position returns the world-space position of the center of mass of the
// specified body.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) Position(id BodyID) dprec.Vec3 {
	body := v.resolve(id, true)
	return body.position
}

// SetPosition changes the world-space position of the center of mass of
// the specified body. The body's collision shapes are moved along
// immediately, so segment queries observe the new position right away.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) SetPosition(id BodyID, position dprec.Vec3) {
	body := v.resolve(id, true)
	body.position = position
	v.refreshPlacement(body)
}

// Rotation returns the world-space orientation of the specified body.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) Rotation(id BodyID) dprec.Quat {
	body := v.resolve(id, true)
	return body.rotation
}

// SetRotation changes the world-space orientation of the specified body.
// The rotation must be a unit quaternion. The body's collision shapes are
// rotated along immediately, so segment queries observe the new
// orientation right away.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) SetRotation(id BodyID, rotation dprec.Quat) {
	body := v.resolve(id, true)
	body.rotation = rotation
	body.recalculateInertia()
	v.refreshPlacement(body)
}

// AttachCollisionSphere attaches a sphere collision shape to the specified
// body and returns an ID through which it can later be detached. The
// sphere is specified in the body's local space and moves along with the
// body.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) AttachCollisionSphere(id BodyID, col CollisionSphere) BodyCollisionShapeID {
	body := v.resolve(id, true)
	shapeID := v.scene.collisionScene.AttachSphere(body.objectID, placement3d.SphereInfo[shapeData]{
		Sphere:    col.Shape,
		Filtering: col.Filtering,
		UserData: shapeData{
			frictionCoefficient:    col.FrictionCoefficient,
			restitutionCoefficient: col.RestitutionCoefficient,
		},
	})
	return BodyCollisionShapeID{
		bodyID:  id,
		shapeID: shapeID,
	}
}

// AttachCollisionBox attaches a box collision shape to the specified body
// and returns an ID through which it can later be detached. The box is
// specified in the body's local space and moves along with the body.
//
// It panics if the ID does not reference a valid body.
func (v BodyView) AttachCollisionBox(id BodyID, col CollisionBox) BodyCollisionShapeID {
	body := v.resolve(id, true)
	shapeID := v.scene.collisionScene.AttachBox(body.objectID, placement3d.BoxInfo[shapeData]{
		Box:       col.Shape,
		Filtering: col.Filtering,
		UserData: shapeData{
			frictionCoefficient:    col.FrictionCoefficient,
			restitutionCoefficient: col.RestitutionCoefficient,
		},
	})
	return BodyCollisionShapeID{
		bodyID:  id,
		shapeID: shapeID,
	}
}

// DetachCollisionShape removes a collision shape that was previously
// attached to the specified body through [BodyView.AttachCollisionSphere]
// or [BodyView.AttachCollisionBox].
//
// It panics if shapeID was not obtained for this same body.
func (v BodyView) DetachCollisionShape(id BodyID, shapeID BodyCollisionShapeID) {
	if id != shapeID.bodyID {
		panic("invalid shape ID for body")
	}
	v.scene.collisionScene.DeleteObjectShape(shapeID.shapeID)
}

func (v BodyView) refreshPlacement(body *bodyState) {
	v.scene.collisionScene.SetObjectTransform(body.objectID, shape3d.Transform{
		Translation: body.position,
		Rotation:    shape3d.RotationFromQuat(body.rotation),
	})
}

func (v BodyView) idFromIndex(index int32) BodyID {
	body := &v.scene.bodies[index]
	return BodyID{
		index:    index,
		revision: body.revision,
	}
}

func (v BodyView) resolve(id BodyID, required bool) *bodyState {
	if id.revision == 0 {
		if required {
			panic("invalid body ID")
		}
		return nil
	}
	body := &v.scene.bodies[id.index]
	if body.revision != id.revision {
		if required {
			panic("invalid body ID")
		}
		return nil
	}
	return body
}

// BodyHandle is an object-oriented alternative to [BodyView] that is
// bound to a specific [BodyID].
//
// It is obtained through [BodyView.Handle].
//
// Wrapping an ID this way needs no allocation of its own, but unlike a
// plain [BodyID] (which holds no pointers), a Handle keeps an internal
// reference to the owning [Scene]. This keeps that Scene reachable for as
// long as the handle is retained, and adds a pointer that the garbage
// collector must trace wherever the handle is stored. Prefer storing IDs
// over handles in long-lived collections unless the convenience is worth
// that cost.
type BodyHandle struct {
	view BodyView
	id   BodyID
}

// ID returns the [BodyID] wrapped by this handle.
func (h BodyHandle) ID() BodyID {
	return h.id
}

// Delete removes the wrapped body.
//
// See [BodyView.Delete] for further details.
func (h BodyHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether the wrapped body has not been deleted.
func (h BodyHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// Mass returns the mass of the wrapped body, in kg.
//
// See [BodyView.Mass] for further details.
func (h BodyHandle) Mass() float64 {
	return h.view.Mass(h.id)
}

// SetMass changes the mass of the wrapped body, in kg.
//
// See [BodyView.SetMass] for further details.
func (h BodyHandle) SetMass(mass float64) {
	h.view.SetMass(h.id, mass)
}

// MomentOfInertia returns the local-space moment of inertia tensor of the
// wrapped body.
//
// See [BodyView.MomentOfInertia] for further details.
func (h BodyHandle) MomentOfInertia() dprec.Mat3 {
	return h.view.MomentOfInertia(h.id)
}

// SetMomentOfInertia changes the local-space moment of inertia tensor of
// the wrapped body.
//
// See [BodyView.SetMomentOfInertia] for further details.
func (h BodyHandle) SetMomentOfInertia(inertia dprec.Mat3) {
	h.view.SetMomentOfInertia(h.id, inertia)
}

// Velocity returns the world-space linear velocity of the wrapped body.
//
// See [BodyView.Velocity] for further details.
func (h BodyHandle) Velocity() dprec.Vec3 {
	return h.view.Velocity(h.id)
}

// SetVelocity changes the world-space linear velocity of the wrapped body.
//
// See [BodyView.SetVelocity] for further details.
func (h BodyHandle) SetVelocity(velocity dprec.Vec3) {
	h.view.SetVelocity(h.id, velocity)
}

// AngularVelocity returns the world-space angular velocity of the wrapped
// body.
//
// See [BodyView.AngularVelocity] for further details.
func (h BodyHandle) AngularVelocity() dprec.Vec3 {
	return h.view.AngularVelocity(h.id)
}

// SetAngularVelocity changes the world-space angular velocity of the
// wrapped body.
//
// See [BodyView.SetAngularVelocity] for further details.
func (h BodyHandle) SetAngularVelocity(angularVelocity dprec.Vec3) {
	h.view.SetAngularVelocity(h.id, angularVelocity)
}

// Position returns the world-space position of the wrapped body.
//
// See [BodyView.Position] for further details.
func (h BodyHandle) Position() dprec.Vec3 {
	return h.view.Position(h.id)
}

// SetPosition changes the world-space position of the wrapped body.
//
// See [BodyView.SetPosition] for further details.
func (h BodyHandle) SetPosition(position dprec.Vec3) {
	h.view.SetPosition(h.id, position)
}

// Rotation returns the world-space orientation of the wrapped body.
//
// See [BodyView.Rotation] for further details.
func (h BodyHandle) Rotation() dprec.Quat {
	return h.view.Rotation(h.id)
}

// SetRotation changes the world-space orientation of the wrapped body.
//
// See [BodyView.SetRotation] for further details.
func (h BodyHandle) SetRotation(rotation dprec.Quat) {
	h.view.SetRotation(h.id, rotation)
}

// AttachCollisionSphere attaches a sphere collision shape to the wrapped
// body.
//
// See [BodyView.AttachCollisionSphere] for further details.
func (h BodyHandle) AttachCollisionSphere(shape CollisionSphere) BodyCollisionShapeID {
	return h.view.AttachCollisionSphere(h.id, shape)
}

// AttachCollisionBox attaches a box collision shape to the wrapped body.
//
// See [BodyView.AttachCollisionBox] for further details.
func (h BodyHandle) AttachCollisionBox(shape CollisionBox) BodyCollisionShapeID {
	return h.view.AttachCollisionBox(h.id, shape)
}

// DetachCollisionShape removes a collision shape from the wrapped body.
//
// See [BodyView.DetachCollisionShape] for further details.
func (h BodyHandle) DetachCollisionShape(shapeID BodyCollisionShapeID) {
	h.view.DetachCollisionShape(h.id, shapeID)
}

type bodyState struct {
	objectID placement3d.ObjectID

	revision                  int32
	firstBodyAcceleratorIndex int32
	firstSoloConstraintIndex  int32
	firstPairConstraintIndex  int32

	invInertiaLocal dprec.Mat3
	invInertia      dprec.Mat3
	invMass         float64

	linearAcceleration  dprec.Vec3
	angularAcceleration dprec.Vec3

	linearVelocity  dprec.Vec3
	angularVelocity dprec.Vec3

	position dprec.Vec3
	rotation dprec.Quat
}

func (s bodyState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}

func (s *bodyState) mass() float64 {
	return 1.0 / s.invMass
}

func (s *bodyState) inertia() dprec.Mat3 {
	return dprec.InverseMat3(s.invInertia)
}

func (b *bodyState) recalculateInertia() {
	b.invInertia = RotatedMomentOfInertia(b.invInertiaLocal, b.rotation)
}

func (b *bodyState) addLinearAcceleration(amount dprec.Vec3) {
	b.linearAcceleration = dprec.Vec3Sum(b.linearAcceleration, amount)
}

func (b *bodyState) addAngularAcceleration(amount dprec.Vec3) {
	b.angularAcceleration = dprec.Vec3Sum(b.angularAcceleration, amount)
}

func (b *bodyState) clampLinearAcceleration(max float64) {
	if b.linearAcceleration.SqrLength() > max*max {
		b.linearAcceleration = dprec.ResizedVec3(b.linearAcceleration, max)
	}
}

func (b *bodyState) clampAngularAcceleration(max float64) {
	if b.angularAcceleration.SqrLength() > max*max {
		b.angularAcceleration = dprec.ResizedVec3(b.angularAcceleration, max)
	}
}

func (b *bodyState) applyForce(force dprec.Vec3) {
	b.addLinearAcceleration(dprec.Vec3Prod(force, b.invMass))
}

func (b *bodyState) applyTorque(torque dprec.Vec3) {
	b.addAngularAcceleration(dprec.Mat3Vec3Prod(b.invInertia, torque))
}

func (b *bodyState) applyOffsetForce(offset, force dprec.Vec3) {
	b.applyForce(force)
	b.applyTorque(dprec.Vec3Cross(offset, force))
}

func (b *bodyState) addLinearVelocity(amount dprec.Vec3) {
	b.linearVelocity = dprec.Vec3Sum(b.linearVelocity, amount)
}

func (b *bodyState) addAngularVelocity(amount dprec.Vec3) {
	b.angularVelocity = dprec.Vec3Sum(b.angularVelocity, amount)
}

func (b *bodyState) clampLinearVelocity(max float64) {
	if b.linearVelocity.SqrLength() > max*max {
		b.linearVelocity = dprec.ResizedVec3(b.linearVelocity, max)
	}
}

func (b *bodyState) clampAngularVelocity(max float64) {
	if b.angularVelocity.SqrLength() > max*max {
		b.angularVelocity = dprec.ResizedVec3(b.angularVelocity, max)
	}
}

func (b *bodyState) translate(offset dprec.Vec3) {
	b.position = dprec.Vec3Sum(b.position, offset)
}

func (b *bodyState) rotate(quat dprec.Quat) {
	b.rotation = dprec.UnitQuat(dprec.QuatProd(quat, b.rotation))
}
