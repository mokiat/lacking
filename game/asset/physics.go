package asset

type BodyDefinition struct {
	Name                   string
	IsStatic               bool
	Mass                   float64
	MomentOfInertia        [3][3]float64
	RestitutionCoefficient float64
	DragFactor             float64
	AngularDragFactor      float64
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
	Translation [3]float64
	Rotation    [4]float64
	Width       float64
	Height      float64
	Lenght      float64
}

type CollisionSphere struct {
	Translation [3]float64
	Rotation    [4]float64
	Radius      float64
}

type CollisionMesh struct {
	Translation [3]float64
	Rotation    [4]float64
	Triangles   []CollisionTriangle
}

type CollisionTriangle struct {
	A [3]float64
	B [3]float64
	C [3]float64
}
