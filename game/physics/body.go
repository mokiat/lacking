package physics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

type BodyID struct {
	index    int32
	revision int32
}

var NilBodyID = BodyID{}

type BodyView struct {
	scene *Scene
}

func (v BodyView) Create(position dprec.Vec3, rotation dprec.Quat) BodyID {
	index := v.scene.allocateBody()

	objectID := v.scene.collisionScene.CreateObject(placement3d.ObjectInfo[bodyData]{
		Position: opt.V(position),
		Rotation: opt.V(rotation),
		UserData: bodyData{
			index: index,
		},
	})

	body := &v.scene.bodies[index]
	body.objectID = objectID
	body.revision++ // progress revision to valid (odd) value
	body.invMass = 1.0
	body.invInertia = dprec.IdentityMat3()
	body.linearVelocity = dprec.ZeroVec3()
	body.angularVelocity = dprec.ZeroVec3()
	body.position = position
	body.rotation = rotation

	return BodyID{
		index:    index,
		revision: body.revision,
	}
}

func (v BodyView) Delete(id BodyID) {
	body := v.resolve(id, true)

	bodyAcceleratorView := v.scene.BodyAccelerators()
	for body.firstBodyAcceleratorIndex != nilIndex {
		bodyAcceleratorView.Delete(bodyAcceleratorView.idFromIndex(body.firstBodyAcceleratorIndex))
	}
	soloConstraintView := v.scene.SoloConstraints()
	for body.firstSoloConstraintIndex != nilIndex {
		soloConstraintView.Delete(soloConstraintView.idFromIndex(body.firstSoloConstraintIndex))
	}
	// TODO: delete pair constraints as well.

	v.scene.collisionScene.DeleteObject(body.objectID)
	body.objectID = placement3d.InvalidObjectID
	body.revision++ // progress revision to invalid (even) value
	v.scene.releaseBody(id.index)
}

func (v BodyView) Handle(id BodyID) BodyHandle {
	return BodyHandle{
		view: v,
		id:   id,
	}
}

func (v BodyView) IsValid(id BodyID) bool {
	body := v.resolve(id, false)
	return body != nil
}

func (v BodyView) Mass(id BodyID) float64 {
	body := v.resolve(id, true)
	return 1.0 / body.invMass
}

func (v BodyView) SetMass(id BodyID, mass float64) {
	body := v.resolve(id, true)
	body.invMass = 1.0 / mass
}

func (v BodyView) MomentOfInertia(id BodyID) dprec.Mat3 {
	body := v.resolve(id, true)
	return dprec.InverseMat3(body.invInertia)
}

func (v BodyView) SetMomentOfInertia(id BodyID, inertia dprec.Mat3) {
	body := v.resolve(id, true)
	body.invInertia = dprec.InverseMat3(inertia)
}

func (v BodyView) Velocity(id BodyID) dprec.Vec3 {
	body := v.resolve(id, true)
	return body.linearVelocity
}

func (v BodyView) SetVelocity(id BodyID, velocity dprec.Vec3) {
	body := v.resolve(id, true)
	body.linearVelocity = velocity
}

func (v BodyView) AngularVelocity(id BodyID) dprec.Vec3 {
	body := v.resolve(id, true)
	return body.angularVelocity
}

func (v BodyView) SetAngularVelocity(id BodyID, angularVelocity dprec.Vec3) {
	body := v.resolve(id, true)
	body.angularVelocity = angularVelocity
}

func (v BodyView) Position(id BodyID) dprec.Vec3 {
	body := v.resolve(id, true)
	return body.position
}

func (v BodyView) SetPosition(id BodyID, position dprec.Vec3) {
	body := v.resolve(id, true)
	body.position = position
	v.refreshPlacement(id, body)
}

func (v BodyView) Rotation(id BodyID) dprec.Quat {
	body := v.resolve(id, true)
	return body.rotation
}

func (v BodyView) SetRotation(id BodyID, rotation dprec.Quat) {
	body := v.resolve(id, true)
	body.rotation = rotation
	v.refreshPlacement(id, body)
}

func (v BodyView) AttachCollisionSphere(id BodyID, col CollisionSphere) CollisionShapeID {
	body := v.resolve(id, true)
	shapeID := v.scene.collisionScene.AttachSphere(body.objectID, placement3d.SphereInfo[shapeData]{
		Sphere:    col.Shape,
		Filtering: col.Filtering,
		UserData: shapeData{
			frictionCoefficient:    col.FrictionCoefficient,
			restitutionCoefficient: col.RestitutionCoefficient,
		},
	})
	return CollisionShapeID{
		bodyID:  id,
		shapeID: shapeID,
	}
}

func (v BodyView) AttachCollisionBox(id BodyID, col CollisionBox) CollisionShapeID {
	body := v.resolve(id, true)
	shapeID := v.scene.collisionScene.AttachBox(body.objectID, placement3d.BoxInfo[shapeData]{
		Box:       col.Shape,
		Filtering: col.Filtering,
		UserData: shapeData{
			frictionCoefficient:    col.FrictionCoefficient,
			restitutionCoefficient: col.RestitutionCoefficient,
		},
	})
	return CollisionShapeID{
		bodyID:  id,
		shapeID: shapeID,
	}
}

func (v BodyView) DetachCollisionShape(id BodyID, shapeID CollisionShapeID) {
	if id != shapeID.bodyID {
		panic("invalid shape ID for body")
	}
	v.scene.collisionScene.DeleteShape(shapeID.shapeID)
}

func (v BodyView) refreshPlacement(id BodyID, body *bodyState) {
	v.scene.collisionScene.SetObjectTransform(body.objectID, shape3d.Transform{
		Translation: body.position,
		Rotation:    shape3d.RotationFromQuat(body.rotation),
	})
}

// func (v BodyView) idFromIndex(index int32) BodyID {
// 	body := &v.scene.bodies[index]
// 	return BodyID{
// 		index:    index,
// 		revision: body.revision,
// 	}
// }

func (v BodyView) resolve(id BodyID, required bool) *bodyState {
	if id.revision == 0 {
		if required {
			panic("invalid global accelerator ID")
		}
		return nil
	}
	body := &v.scene.bodies[id.index]
	if body.revision != id.revision {
		if required {
			panic("invalid global accelerator ID")
		}
		return nil
	}
	return body
}

type BodyHandle struct {
	view BodyView
	id   BodyID
}

func (h BodyHandle) ID() BodyID {
	return h.id
}

func (h BodyHandle) Delete() {
	h.view.Delete(h.id)
}

func (h BodyHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

func (h BodyHandle) Mass() float64 {
	return h.view.Mass(h.id)
}

func (h BodyHandle) SetMass(mass float64) {
	h.view.SetMass(h.id, mass)
}

func (h BodyHandle) MomentOfInertia() dprec.Mat3 {
	return h.view.MomentOfInertia(h.id)
}

func (h BodyHandle) SetMomentOfInertia(inertia dprec.Mat3) {
	h.view.SetMomentOfInertia(h.id, inertia)
}

func (h BodyHandle) Velocity() dprec.Vec3 {
	return h.view.Velocity(h.id)
}

func (h BodyHandle) SetVelocity(velocity dprec.Vec3) {
	h.view.SetVelocity(h.id, velocity)
}

func (h BodyHandle) AngularVelocity() dprec.Vec3 {
	return h.view.AngularVelocity(h.id)
}

func (h BodyHandle) SetAngularVelocity(angularVelocity dprec.Vec3) {
	h.view.SetAngularVelocity(h.id, angularVelocity)
}

func (h BodyHandle) Position() dprec.Vec3 {
	return h.view.Position(h.id)
}

func (h BodyHandle) SetPosition(position dprec.Vec3) {
	h.view.SetPosition(h.id, position)
}

func (h BodyHandle) Rotation() dprec.Quat {
	return h.view.Rotation(h.id)
}

func (h BodyHandle) SetRotation(rotation dprec.Quat) {
	h.view.SetRotation(h.id, rotation)
}

func (h BodyHandle) AttachCollisionSphere(shape CollisionSphere) CollisionShapeID {
	return h.view.AttachCollisionSphere(h.id, shape)
}

func (h BodyHandle) AttachCollisionBox(shape CollisionBox) CollisionShapeID {
	return h.view.AttachCollisionBox(h.id, shape)
}

func (h BodyHandle) DetachCollisionShape(shapeID CollisionShapeID) {
	h.view.DetachCollisionShape(h.id, shapeID)
}

type CollisionShapeID struct {
	bodyID  BodyID
	shapeID placement3d.ShapeID
}

type CollisionShape[T any] struct {
	Shape                  T
	FrictionCoefficient    float64
	RestitutionCoefficient float64
	Filtering              placement3d.FilterInfo
}

type CollisionSphere CollisionShape[shape3d.Sphere]

type CollisionBox CollisionShape[shape3d.Box]

type bodyData struct {
	index int32
}

type shapeData struct {
	frictionCoefficient    float64
	restitutionCoefficient float64
}

type bodyState struct {
	objectID placement3d.ObjectID

	revision                  int32
	firstBodyAcceleratorIndex int32
	firstSoloConstraintIndex  int32

	invMass    float64
	invInertia dprec.Mat3

	linearVelocity  dprec.Vec3
	angularVelocity dprec.Vec3

	position dprec.Vec3
	rotation dprec.Quat
}

func (s bodyState) IsActive() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}

func (b *bodyState) AddVelocity(amount dprec.Vec3) {
	b.linearVelocity = dprec.Vec3Sum(b.linearVelocity, amount)
}

func (b *bodyState) AddAngularVelocity(amount dprec.Vec3) {
	b.angularVelocity = dprec.Vec3Sum(b.angularVelocity, amount)
}

func (b *bodyState) ClampVelocity(max float64) {
	if b.linearVelocity.SqrLength() > max*max {
		b.linearVelocity = dprec.ResizedVec3(b.linearVelocity, max)
	}
}

func (b *bodyState) ClampAngularVelocity(max float64) {
	if b.angularVelocity.SqrLength() > max*max {
		b.angularVelocity = dprec.ResizedVec3(b.angularVelocity, max)
	}
}

func (b *bodyState) Translate(offset dprec.Vec3) {
	b.position = dprec.Vec3Sum(b.position, offset)
}

func (b *bodyState) VectorRotate(vector dprec.Vec3) {
	const angularEpsilon = float64(0.00001)
	if radians := vector.Length(); dprec.Abs(radians) > angularEpsilon {
		b.Rotate(dprec.RotationQuat(dprec.Radians(radians), vector))
	}
}

func (b *bodyState) Rotate(quat dprec.Quat) {
	b.rotation = dprec.UnitQuat(dprec.QuatProd(quat, b.rotation))
}
