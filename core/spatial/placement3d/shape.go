package placement3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d"
	"github.com/mokiat/lacking/core/spatial/query3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// InvalidShapeID indicates a shape that can never be part of the scene.
const InvalidShapeID = ShapeID(nilIndex)

// ShapeID is a reference to a shape in the scene.
type ShapeID int32

// SphereInfo contains the information needed to create a sphere shape.
type SphereInfo[S any] struct {

	// Filtering holds the collision-filtering metadata for the shape.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the shape.
	UserData S

	// Sphere contains the sphere information.
	Sphere shape3d.Sphere
}

// BoxInfo contains the information needed to create a box shape.
type BoxInfo[S any] struct {

	// Filtering holds the collision-filtering metadata for the shape.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the shape.
	UserData S

	// Box contains the box information.
	Box shape3d.Box
}

type shape[S any] struct {
	objectIndex    int32
	nextShapeIndex int32
	prevShapeIndex int32
	spatialID      query3d.TreeItemID
	filterRepresentation
	shapeRepresentation
	userData S
}

func shapesCanIntersect[S any](a, b *shape[S]) bool {
	if a.objectIndex >= b.objectIndex {
		return false // prevent self-intersection and repeated checks
	}
	return a.filterRepresentation.canInteractWith(&b.filterRepresentation)
}

type shapeRepresentation struct {
	lsBSphere shape3d.Sphere
	wsBSphere shape3d.Sphere

	lsTransform shape3d.Transform
	wsTransform shape3d.Transform

	kind       shapeKind
	points     []dprec.Vec3
	skinRadius float64
}

func (s *shapeRepresentation) update(parentTransform shape3d.Transform) {
	s.wsBSphere = shape3d.TransformedSphere(s.lsBSphere, parentTransform)

	s.wsTransform = shape3d.ChainedTransform(
		parentTransform,
		s.lsTransform,
	)
}

func (s *shapeRepresentation) boundingSphere() shape3d.Sphere {
	return s.wsBSphere
}

func (s *shapeRepresentation) gjkShape() gjk3d.Shape {
	return gjk3d.Shape{
		Position:   s.wsTransform.Translation,
		Rotation:   s.wsTransform.Rotation,
		Points:     s.points,
		SkinRadius: s.skinRadius,
	}
}

func (s *shapeRepresentation) toSphere() shape3d.Sphere {
	return shape3d.Sphere{
		Center: s.wsTransform.Translation,
		Radius: s.skinRadius,
	}
}

func (s *shapeRepresentation) toBox() shape3d.Box {
	var halfWidth, halfHeight, halfLength float64
	for _, point := range s.points {
		halfWidth = max(halfWidth, point.X)
		halfHeight = max(halfHeight, point.Y)
		halfLength = max(halfLength, point.Z)
	}
	return shape3d.Box{
		Center:     s.wsTransform.Translation,
		Rotation:   s.wsTransform.Rotation,
		HalfWidth:  halfWidth,
		HalfHeight: halfHeight,
		HalfLength: halfLength,
	}
}

type shapeKind uint32

const (
	shapeKindSphere shapeKind = iota
	shapeKindBox
	shapeKindCapsule
	shapeKindConvexHull
)
