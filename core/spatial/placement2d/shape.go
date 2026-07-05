package placement2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d"
	"github.com/mokiat/lacking/core/spatial/query2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// InvalidShapeID indicates a shape that can never be part of the scene.
const InvalidShapeID = ShapeID(nilIndex)

// ShapeID is a reference to a shape in the scene.
type ShapeID int32

// CircleInfo contains the information needed to create a circle shape.
type CircleInfo[S any] struct {

	// Filtering holds the collision-filtering metadata for the shape.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the shape.
	UserData S

	// Circle contains the circle information.
	Circle shape2d.Circle
}

// RectangleInfo contains the information needed to create a rectangle shape.
type RectangleInfo[S any] struct {

	// Filtering holds the collision-filtering metadata for the shape.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the shape.
	UserData S

	// Rectangle contains the rectangle information.
	Rectangle shape2d.Rectangle
}

type shape[S any] struct {
	objectIndex    int32
	nextShapeIndex int32
	prevShapeIndex int32
	spatialID      query2d.TreeItemID
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
	lsBCircle shape2d.Circle
	wsBCircle shape2d.Circle

	lsTransform shape2d.Transform
	wsTransform shape2d.Transform

	kind       shapeKind
	points     []dprec.Vec2
	skinRadius float64
}

func (s *shapeRepresentation) update(parentTransform shape2d.Transform) {
	s.wsBCircle = shape2d.TransformedCircle(s.lsBCircle, parentTransform)

	s.wsTransform = shape2d.ChainedTransform(
		parentTransform,
		s.lsTransform,
	)
}

func (s *shapeRepresentation) boundingCircle() shape2d.Circle {
	return s.wsBCircle
}

func (s *shapeRepresentation) gjkShape() gjk2d.Shape {
	return gjk2d.Shape{
		Position:   s.wsTransform.Translation,
		Rotation:   s.wsTransform.Rotation,
		Points:     s.points,
		SkinRadius: s.skinRadius,
	}
}

func (s *shapeRepresentation) toCircle() shape2d.Circle {
	return shape2d.Circle{
		Center: s.wsTransform.Translation,
		Radius: s.skinRadius,
	}
}

func (s *shapeRepresentation) toRectangle() shape2d.Rectangle {
	var halfWidth, halfHeight float64
	for _, point := range s.points {
		halfWidth = max(halfWidth, point.X)
		halfHeight = max(halfHeight, point.Y)
	}
	return shape2d.Rectangle{
		Center:     s.wsTransform.Translation,
		Rotation:   s.wsTransform.Rotation,
		HalfWidth:  halfWidth,
		HalfHeight: halfHeight,
	}
}

type shapeKind uint32

const (
	shapeKindCircle shapeKind = iota
	shapeKindRectangle
	shapeKindCapsule
	shapeKindConvexHull
)
