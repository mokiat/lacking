package asset

import "github.com/mokiat/gomath/dprec"

type BodyDefinition struct {
	Name                   string
	Mass                   float64
	MomentOfInertia        dprec.Mat3
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
	Translation dprec.Vec3
	Rotation    dprec.Quat
	Width       float64
	Height      float64
	Lenght      float64
}

type CollisionSphere struct {
	Translation dprec.Vec3
	Rotation    dprec.Quat
	Radius      float64
}

type CollisionMesh struct {
	Translation dprec.Vec3
	Rotation    dprec.Quat
	Triangles   []CollisionTriangle
}

type CollisionTriangle struct {
	A dprec.Vec3
	B dprec.Vec3
	C dprec.Vec3
}
