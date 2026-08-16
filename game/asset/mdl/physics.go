package mdl

import "github.com/mokiat/gomath/dprec"

type PhysicsBody struct {
	*Object
	mass             float64
	momentOfInertia  dprec.Mat3
	collisionSpheres []*CollisionSphere
	collisionBoxes   []*CollisionBox
}

func NewPhysicsBody() *PhysicsBody {
	return &PhysicsBody{
		Object:           NewObject(),
		mass:             1.0,
		momentOfInertia:  dprec.IdentityMat3(),
		collisionSpheres: []*CollisionSphere{},
		collisionBoxes:   []*CollisionBox{},
	}
}

func (b *PhysicsBody) Mass() float64 {
	return b.mass
}

func (b *PhysicsBody) SetMass(value float64) {
	b.mass = value
}

func (b *PhysicsBody) MomentOfInertia() dprec.Mat3 {
	return b.momentOfInertia
}

func (b *PhysicsBody) SetMomentOfInertia(value dprec.Mat3) {
	b.momentOfInertia = value
}

func (b *PhysicsBody) CollisionSpheres() []*CollisionSphere {
	return b.collisionSpheres
}

func (b *PhysicsBody) AddCollisionSphere(value *CollisionSphere) {
	b.collisionSpheres = append(b.collisionSpheres, value)
}

func (b *PhysicsBody) CollisionBoxes() []*CollisionBox {
	return b.collisionBoxes
}

func (b *PhysicsBody) AddCollisionBox(value *CollisionBox) {
	b.collisionBoxes = append(b.collisionBoxes, value)
}

type PhysicsTerrain struct {
	*Object
	collisionMeshes []*CollisionMesh
}

func NewPhysicsTerrain() *PhysicsTerrain {
	return &PhysicsTerrain{
		Object:          NewObject(),
		collisionMeshes: []*CollisionMesh{},
	}
}

func (t *PhysicsTerrain) CollisionMeshes() []*CollisionMesh {
	return t.collisionMeshes
}

func (t *PhysicsTerrain) SetCollisionMeshes(collisionMeshes []*CollisionMesh) {
	t.collisionMeshes = collisionMeshes
}

func (t *PhysicsTerrain) AddCollisionMesh(value *CollisionMesh) {
	t.collisionMeshes = append(t.collisionMeshes, value)
}

func (t *PhysicsTerrain) AddCollisionMeshes(collisionMeshes []*CollisionMesh) {
	t.collisionMeshes = append(t.collisionMeshes, collisionMeshes...)
}

type CollisionShape struct {
	frictionCoefficient    float64
	restitutionCoefficient float64
}

func (s *CollisionShape) FrictionCoefficient() float64 {
	return s.frictionCoefficient
}

func (s *CollisionShape) SetFrictionCoefficient(value float64) {
	s.frictionCoefficient = value
}

func (s *CollisionShape) RestitutionCoefficient() float64 {
	return s.restitutionCoefficient
}

func (s *CollisionShape) SetRestitutionCoefficient(value float64) {
	s.restitutionCoefficient = value
}

type CollisionSphere struct {
	CollisionShape
	translation dprec.Vec3
	radius      float64
}

func NewCollisionSphere() *CollisionSphere {
	return &CollisionSphere{
		translation: dprec.ZeroVec3(),
	}
}

func (s *CollisionSphere) Translation() dprec.Vec3 {
	return s.translation
}

func (s *CollisionSphere) SetTranslation(value dprec.Vec3) {
	s.translation = value
}

func (s *CollisionSphere) Radius() float64 {
	return s.radius
}

func (s *CollisionSphere) SetRadius(value float64) {
	s.radius = value
}

type CollisionBox struct {
	CollisionShape
	translation dprec.Vec3
	rotation    dprec.Quat
	width       float64
	height      float64
	length      float64
}

func NewCollisionBox() *CollisionBox {
	return &CollisionBox{
		translation: dprec.ZeroVec3(),
		rotation:    dprec.IdentityQuat(),
	}
}

func (b *CollisionBox) Translation() dprec.Vec3 {
	return b.translation
}

func (b *CollisionBox) SetTranslation(value dprec.Vec3) {
	b.translation = value
}

func (b *CollisionBox) Rotation() dprec.Quat {
	return b.rotation
}

func (b *CollisionBox) SetRotation(value dprec.Quat) {
	b.rotation = value
}

func (b *CollisionBox) Width() float64 {
	return b.width
}

func (b *CollisionBox) SetWidth(value float64) {
	b.width = value
}

func (b *CollisionBox) Height() float64 {
	return b.height
}

func (b *CollisionBox) SetHeight(value float64) {
	b.height = value
}

func (b *CollisionBox) Length() float64 {
	return b.length
}

func (b *CollisionBox) SetLength(value float64) {
	b.length = value
}

type CollisionMesh struct {
	CollisionShape
	translation dprec.Vec3
	rotation    dprec.Quat
	triangles   []CollisionTriangle
}

func NewCollisionMesh() *CollisionMesh {
	return &CollisionMesh{
		translation: dprec.ZeroVec3(),
		rotation:    dprec.IdentityQuat(),
	}
}

func (m *CollisionMesh) Translation() dprec.Vec3 {
	return m.translation
}

func (m *CollisionMesh) SetTranslation(value dprec.Vec3) {
	m.translation = value
}

func (m *CollisionMesh) Rotation() dprec.Quat {
	return m.rotation
}

func (m *CollisionMesh) SetRotation(value dprec.Quat) {
	m.rotation = value
}

func (m *CollisionMesh) Triangles() []CollisionTriangle {
	return m.triangles
}

func (m *CollisionMesh) SetTriangles(triangles []CollisionTriangle) {
	m.triangles = triangles
}

func (m *CollisionMesh) AddTriangle(value CollisionTriangle) {
	m.triangles = append(m.triangles, value)
}

type CollisionTriangle struct {
	A dprec.Vec3
	B dprec.Vec3
	C dprec.Vec3
}

// TODO: Move somewhere else.
type Placed[T any] struct {
	Node  *Node
	Value T
}
