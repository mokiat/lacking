package placement3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// ShapeID is a reference to a shape in the scene.
type ShapeID shapeRef

// ShapeInfo contains information needed to create a new shape in the scene.
type ShapeInfo[S any] struct {

	// RejectGroup becomes active if a value larger than zero is specified.
	// Shapes that share the same reject group are not checked for intersection.
	RejectGroup uint32

	// SourceMask specifies the layers in which this shape is positioned.
	SourceMask opt.T[uint32]

	// TargetMask specifies the layers with which this shape can intersect.
	TargetMask opt.T[uint32]

	// UserData allows one to attach custom user data to a shape.
	UserData S
}

// SphereInfo contains the information needed to create a sphere shape.
type SphereInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Sphere contains the sphere information.
	Sphere shape3d.Sphere
}

// BoxInfo contains the information needed to create a box shape.
type BoxInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Box contains the box information.
	Box shape3d.Box
}

// MeshInfo contains the information needed to create a mesh shape.
type MeshInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Mesh contains the mesh information.
	Mesh shape3d.Mesh
}
