package mdl

import "github.com/mokiat/gomath/dprec"

func NewBodyMaterial() *BodyMaterial {
	return &BodyMaterial{
		frictionCoefficient:    1.0,
		restitutionCoefficient: 0.5,
	}
}

type BodyMaterial struct {
	frictionCoefficient    float64
	restitutionCoefficient float64
}

func (m *BodyMaterial) FrictionCoefficient() float64 {
	return m.frictionCoefficient
}

func (m *BodyMaterial) SetFrictionCoefficient(value float64) {
	m.frictionCoefficient = value
}

func (m *BodyMaterial) RestitutionCoefficient() float64 {
	return m.restitutionCoefficient
}

func (m *BodyMaterial) SetRestitutionCoefficient(value float64) {
	m.restitutionCoefficient = value
}

func NewBodyDefinition(material *BodyMaterial) *BodyDefinition {
	return &BodyDefinition{
		material: material,
	}
}

type BodyDefinition struct {
	material          *BodyMaterial
	mass              float64
	momentOfInertia   dprec.Mat3
	dragFactor        float64
	angularDragFactor float64
	collisionBoxes    []*CollisionBox
	collisionSpheres  []*CollisionSphere
	collisionMeshes   []*CollisionMesh
}

func (d *BodyDefinition) Material() *BodyMaterial {
	return d.material
}

func (d *BodyDefinition) Mass() float64 {
	return d.mass
}

func (d *BodyDefinition) SetMass(value float64) {
	d.mass = value
}

func (d *BodyDefinition) MomentOfInertia() dprec.Mat3 {
	return d.momentOfInertia
}

func (d *BodyDefinition) SetMomentOfInertia(value dprec.Mat3) {
	d.momentOfInertia = value
}

func (d *BodyDefinition) DragFactor() float64 {
	return d.dragFactor
}

func (d *BodyDefinition) SetDragFactor(value float64) {
	d.dragFactor = value
}

func (d *BodyDefinition) AngularDragFactor() float64 {
	return d.angularDragFactor
}

func (d *BodyDefinition) SetAngularDragFactor(value float64) {
	d.angularDragFactor = value
}

func (d *BodyDefinition) CollisionBoxes() []*CollisionBox {
	return d.collisionBoxes
}

func (d *BodyDefinition) AddCollisionBox(value *CollisionBox) {
	d.collisionBoxes = append(d.collisionBoxes, value)
}

func (d *BodyDefinition) CollisionSpheres() []*CollisionSphere {
	return d.collisionSpheres
}

func (d *BodyDefinition) AddCollisionSphere(value *CollisionSphere) {
	d.collisionSpheres = append(d.collisionSpheres, value)
}

func (d *BodyDefinition) CollisionMeshes() []*CollisionMesh {
	return d.collisionMeshes
}

func (d *BodyDefinition) SetCollisionMeshes(collisionMeshes []*CollisionMesh) {
	d.collisionMeshes = collisionMeshes
}

func (d *BodyDefinition) AddCollisionMesh(value *CollisionMesh) {
	d.collisionMeshes = append(d.collisionMeshes, value)
}

func (d *BodyDefinition) AddCollisionMeshes(collisionMeshes []*CollisionMesh) {
	d.collisionMeshes = append(d.collisionMeshes, collisionMeshes...)
}

func NewCollisionBox() *CollisionBox {
	return &CollisionBox{
		translation: dprec.ZeroVec3(),
		rotation:    dprec.IdentityQuat(),
	}
}

type CollisionBox struct {
	translation dprec.Vec3
	rotation    dprec.Quat
	width       float64
	height      float64
	length      float64
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

func NewCollisionSphere() *CollisionSphere {
	return &CollisionSphere{
		translation: dprec.ZeroVec3(),
	}
}

type CollisionSphere struct {
	translation dprec.Vec3
	radius      float64
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

func NewCollisionMesh() *CollisionMesh {
	return &CollisionMesh{
		translation: dprec.ZeroVec3(),
		rotation:    dprec.IdentityQuat(),
	}
}

type CollisionMesh struct {
	translation dprec.Vec3
	rotation    dprec.Quat
	triangles   []CollisionTriangle
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

func NewBody(definition *BodyDefinition) *Body {
	return &Body{
		definition: definition,
	}
}

type Body struct {
	definition *BodyDefinition
}

func (b *Body) Definition() *BodyDefinition {
	return b.definition
}
