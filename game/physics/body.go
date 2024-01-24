package physics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/game/physics/solver"
	"github.com/mokiat/lacking/util/spatial"
)

var invalidBodyState = &bodyState{}

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

// Body represents a physical body that has physics
// act upon it.
type Body struct {
	scene     *Scene
	reference indexReference
}

// // Name returns the name of this body.
func (b Body) Name() string {
	state := b.state()
	return state.name
}

// SetName sets a new name for this body.
func (b Body) SetName(name string) {
	state := b.state()
	state.name = name
}

// Mass returns the mass of this body in kg.
func (b Body) Mass() float64 {
	state := b.state()
	return state.mass
}

// SetMass changes the mass of this body.
func (b Body) SetMass(mass float64) {
	state := b.state()
	state.mass = mass
}

// MomentOfInertia returns the moment of inertia, or
// rotational inertia of this body.
func (b Body) MomentOfInertia() dprec.Mat3 {
	state := b.state()
	return state.momentOfInertia
}

// SetMomentOfInertia changes the moment of inertia
// of this body.
func (b Body) SetMomentOfInertia(inertia dprec.Mat3) {
	state := b.state()
	state.momentOfInertia = inertia
}

// // RestitutionCoefficient returns the restitution
// // coefficient of this body. Valid values are in
// // the range [0.0 - 1.0], where 0.0 means that the
// // body does not bounce and 1.0 means that it bounds
// // back with the same velocity. In reality the amount
// // that the body will bounce depends on the restitution
// // coefficients of both bodies colliding. Furthermore,
// // due to computational errors, the bounce will eventually
// // stop.
// func (b *Body) RestitutionCoefficient() float64 {
// 	return b.restitutionCoefficient
// }

// // SetRestitutionCoefficient changes the restitution
// // coefficient for this body.
// func (b *Body) SetRestitutionCoefficient(coefficient float64) {
// 	b.restitutionCoefficient = coefficient
// }

// // DragCoefficient returns the drag factor of this body.
// func (b *Body) DragFactor() float64 {
// 	return b.dragFactor
// }

// // SetDragFactor sets the drag factor for this body.
// // The drag factor is the drag coefficient multiplied
// // by the area and divided in half.
// func (b *Body) SetDragFactor(factor float64) {
// 	b.dragFactor = factor
// }

// // AngularDragFactor returns the angular drag factor
// // for this body.
// func (b *Body) AngularDragFactor() float64 {
// 	return b.angularDragFactor
// }

// // SetAngularDragFactor sets the angular factor for this body.
// // The angular factor is similar to the drag factor, except
// // that it deals with the drag induced by the rotation of
// // the body.
// func (b *Body) SetAngularDragFactor(factor float64) {
// 	b.angularDragFactor = factor
// }

// Position returns the body's position in world
// space.
func (b Body) Position() dprec.Vec3 {
	state := b.state()
	return state.position
}

// SetPosition changes the position of this body.
func (b Body) SetPosition(position dprec.Vec3) {
	state := b.state()
	state.position = position
	state.intermediatePosition = position

	// TODO
	// b.scene.bodyOctree.Update(state.itemID, state.position, state.bsRadius)
	// TODO: Do this only on demand.
	// b.invalidateCollisionShapes()
}

// IntermediatePosition returns the position of the Body as would
// be seen by the current frame.
//
// NOTE: The physics engine can advance past the current frame,
// which is the reason for this method.
func (b Body) IntermediatePosition() dprec.Vec3 {
	state := b.state()
	return state.intermediatePosition
}

// Rotation returns the quaternion rotation
// of this body.
func (b Body) Rotation() dprec.Quat {
	state := b.state()
	return state.rotation
}

// SetRotation changes the quaterntion rotation
// of this body.
func (b Body) SetRotation(rotation dprec.Quat) {
	state := b.state()
	state.rotation = rotation
	state.intermediateRotation = rotation
}

// IntermediateRotation returns the rotation of the Body as would
// be seen by the current frame.
//
// Note: The physics engine can advance past the current frame,
// which is the reason for this method.
func (b Body) IntermediateRotation() dprec.Quat {
	state := b.state()
	return state.intermediateRotation
}

// Velocity returns the velocity of this body.
func (b Body) Velocity() dprec.Vec3 {
	state := b.state()
	return state.velocity
}

// SetVelocity changes the velocity of this body.
func (b Body) SetVelocity(velocity dprec.Vec3) {
	state := b.state()
	state.velocity = velocity
}

// AngularVelocity returns the angular velocity
// of this body.
func (b Body) AngularVelocity() dprec.Vec3 {
	state := b.state()
	return state.angularVelocity
}

// SetAngularVelocity changes the angular velocity
// of this body.
func (b Body) SetAngularVelocity(angularVelocity dprec.Vec3) {
	state := b.state()
	state.angularVelocity = angularVelocity
}

// CollisionGroup returns the collision group for this body. Two bodies
// with the same collision group are not checked for collisions.
func (b Body) CollisionGroup() int {
	state := b.state()
	return state.collisionGroup
}

// SetCollisionGroup changes the collision group for this body.
//
// A value of 0 disables the collision group.
func (b Body) SetCollisionGroup(group int) {
	state := b.state()
	state.collisionGroup = group
}

// CollisionSet contains the collision shapes for this body.
func (b Body) CollisionSet() collision.Set {
	state := b.state()
	return state.collisionSet
}

// // AerodynamicShapes returns a slice of shapes that
// // dictate how this body is affected by relative air
// // motion.
// func (b *Body) AerodynamicShapes() []AerodynamicShape {
// 	return b.aerodynamicShapes
// }

// // SetAerodynamicShapes sets the aerodynamics shapes
// // to be used when calculating wind drag and lift.
// func (b *Body) SetAerodynamicShapes(shapes []AerodynamicShape) {
// 	b.aerodynamicShapes = shapes
// }

// Delete removes this physical body.
func (b Body) Delete() {
	deleteBody(b.scene, b.reference)
}

func (b Body) state() *bodyState {
	index := b.reference.Index
	state := &b.scene.bodies[index]
	if state.reference != b.reference {
		return invalidBodyState
	}
	return state
}

// func (b *Body) applyOffsetForce(offset, force dprec.Vec3) {
// 	b.applyForce(force)
// 	b.applyTorque(dprec.Vec3Cross(offset, force))
// }

// func (b *Body) applyImpulse(impulse dprec.Vec3) {
// 	b.addVelocity(dprec.Vec3Quot(impulse, b.mass))
// }

// func (b *Body) applyAngularImpulse(impulse dprec.Vec3) {
// 	// FIXME: the moment of intertia is in local space, whereas the impulse is in world space
// 	b.addAngularVelocity(dprec.Mat3Vec3Prod(dprec.InverseMat3(b.momentOfInertia), impulse))
// }

// func (b *Body) applyOffsetImpulse(offset, impulse dprec.Vec3) {
// 	b.applyImpulse(impulse)
// 	b.applyAngularImpulse(dprec.Vec3Cross(offset, impulse))
// }

// func (b *Body) applyNudge(nudge dprec.Vec3) {
// 	b.translate(dprec.Vec3Quot(nudge, b.mass))
// }

// func (b *Body) applyAngularNudge(nudge dprec.Vec3) {
// 	// FIXME: the moment of intertia is in local space, whereas the torque is in world space
// 	b.vectorRotate(dprec.Mat3Vec3Prod(dprec.InverseMat3(b.momentOfInertia), nudge))
// }

// func (b *Body) applyOffsetNudge(offset, nudge dprec.Vec3) {
// 	b.applyNudge(nudge)
// 	b.applyAngularNudge(dprec.Vec3Cross(offset, nudge))
// }

type bodyState struct {
	reference indexReference

	name       string
	definition *BodyDefinition

	itemID spatial.DynamicOctreeItemID

	mass            float64
	momentOfInertia dprec.Mat3

	frictionCoefficient    float64
	restitutionCoefficient float64

	dragFactor        float64
	angularDragFactor float64

	oldPosition dprec.Vec3
	oldRotation dprec.Quat

	position dprec.Vec3
	rotation dprec.Quat

	intermediatePosition dprec.Vec3
	intermediateRotation dprec.Quat

	linearAcceleration  dprec.Vec3
	angularAcceleration dprec.Vec3

	velocity        dprec.Vec3
	angularVelocity dprec.Vec3

	bsRadius          float64
	collisionGroup    int
	collisionSet      collision.Set
	aerodynamicShapes []AerodynamicShape
}

func (s bodyState) IsActive() bool {
	return s.reference.IsValid()
}

func (b *bodyState) InvalidateCollisionShapes(scene *Scene) {
	transform := collision.TRTransform(b.position, b.rotation)
	b.collisionSet.Replace(b.definition.collisionSet, transform)

	bs := b.collisionSet.BoundingSphere()
	delta := dprec.Vec3Diff(bs.Position(), b.position)
	b.bsRadius = delta.Length() + bs.Radius()
	scene.bodyOctree.Update(b.itemID, b.position, b.bsRadius)
}

func (b *bodyState) ResetLinearAcceleration() {
	b.linearAcceleration = dprec.ZeroVec3()
}

func (b *bodyState) ResetAngularAcceleration() {
	b.angularAcceleration = dprec.ZeroVec3()
}

func (b *bodyState) AddLinearAcceleration(amount dprec.Vec3) {
	b.linearAcceleration = dprec.Vec3Sum(b.linearAcceleration, amount)
}

func (b *bodyState) AddAngularAcceleration(amount dprec.Vec3) {
	b.angularAcceleration = dprec.Vec3Sum(b.angularAcceleration, amount)
}

func (b *bodyState) ApplyForce(force dprec.Vec3) {
	b.AddLinearAcceleration(dprec.Vec3Quot(force, b.mass))
}

func (b *bodyState) ApplyTorque(torque dprec.Vec3) {
	// FIXME: the moment of intertia is in local space, whereas the torque is in world space
	b.AddAngularAcceleration(dprec.Mat3Vec3Prod(dprec.InverseMat3(b.momentOfInertia), torque))
}

func (b *bodyState) ClampLinearAcceleration(max float64) {
	if b.linearAcceleration.SqrLength() > max*max {
		b.linearAcceleration = dprec.ResizedVec3(b.linearAcceleration, max)
	}
}

func (b *bodyState) ClampAngularAcceleration(max float64) {
	if b.angularAcceleration.SqrLength() > max*max {
		b.angularAcceleration = dprec.ResizedVec3(b.angularAcceleration, max)
	}
}

func (b *bodyState) AddVelocity(amount dprec.Vec3) {
	b.velocity = dprec.Vec3Sum(b.velocity, amount)
}

func (b *bodyState) AddAngularVelocity(amount dprec.Vec3) {
	b.angularVelocity = dprec.Vec3Sum(b.angularVelocity, amount)
}

func (b *bodyState) ClampVelocity(max float64) {
	if b.velocity.SqrLength() > max*max {
		b.velocity = dprec.ResizedVec3(b.velocity, max)
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

func createBody(scene *Scene, info BodyInfo) Body {
	var freeIndex uint32
	if scene.freeBodyIndices.IsEmpty() {
		freeIndex = uint32(len(scene.bodies))
		scene.bodies = append(scene.bodies, bodyState{})
		scene.bodyAccelerationTargets = append(scene.bodyAccelerationTargets, solver.AccelerationTarget{})
		scene.bodyConstraintPlaceholders = append(scene.bodyConstraintPlaceholders, solver.Placeholder{})
	} else {
		freeIndex = scene.freeBodyIndices.Pop()
	}

	reference := newIndexReference(freeIndex, scene.nextRevision())
	body := bodyState{
		reference: reference,

		name:       info.Name,
		definition: info.Definition,

		itemID: scene.bodyOctree.Insert(
			info.Position, 1.0, freeIndex,
		),

		mass:            info.Definition.mass,
		momentOfInertia: info.Definition.momentOfInertia,

		frictionCoefficient:    info.Definition.frictionCoefficient,
		restitutionCoefficient: info.Definition.restitutionCoefficient,

		dragFactor:        info.Definition.dragFactor,
		angularDragFactor: info.Definition.angularDragFactor,

		position: info.Position,
		rotation: info.Rotation,

		collisionGroup:    info.Definition.collisionGroup,
		aerodynamicShapes: info.Definition.aerodynamicShapes,
	}

	// FIXME
	scene.bodyOctree.Update(body.itemID, body.position, body.bsRadius)
	body.InvalidateCollisionShapes(scene)

	scene.bodies[freeIndex] = body

	return Body{
		scene:     scene,
		reference: reference,
	}
}

func deleteBody(scene *Scene, reference indexReference) {
	index := reference.Index
	state := &scene.bodies[index]
	if state.reference == reference {
		state.reference = newIndexReference(index, 0)

		// TODO

		scene.freeBodyIndices.Push(index)
	}
}
