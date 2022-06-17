package asset

type BodyDefinition struct {
	Name                   string
	IsStatic               bool
	Mass                   float32
	MomentOfInertia        [3][3]float32
	RestitutionCoefficient float32
	DragFactor             float32
	AngularDragFactor      float32
	CollisionBoxes         []CollisionBox
	CollisionSpheres       []CollisionSphere
	CollisionMeshes        []CollisionMesh
}

type BodyInstance struct {
	Name      string
	NodeIndex int32
	BodyIndex int32
}

type CollisionBox struct {
	Translation [3]float32
	Rotation    [4]float32
	Width       float32
	Height      float32
	Lenght      float32
}

type CollisionSphere struct {
	Translation [3]float32
	Rotation    [4]float32
	Radius      float32
}

type CollisionMesh struct {
	Translation [3]float32
	Rotation    [4]float32
	Triangles   []CollisionTriangle
}

type CollisionTriangle struct {
	A [3]float32
	B [3]float32
	C [3]float32
}
