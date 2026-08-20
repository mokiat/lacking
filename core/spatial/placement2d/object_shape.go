package placement2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d"
	"github.com/mokiat/lacking/core/spatial/query2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// NilObjectShapeID indicates an object shape that can never be part of the
// scene.
//
// It is also used to denote the absence of a source shape in contacts that
// were produced by a query primitive rather than by a scene shape.
const NilObjectShapeID = ObjectShapeID(nilIndex)

// ObjectShapeID is a reference to a convex shape that is attached to an object
// in the scene.
type ObjectShapeID int32

// CircleInfo contains the information needed to create a circle shape.
type CircleInfo[S any] struct {

	// Filtering holds the collision-filtering metadata for the shape.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the shape.
	UserData S

	// Circle contains the circle information.
	//
	// It is specified in the local space of the object that the shape is
	// attached to.
	Circle shape2d.Circle
}

// RectangleInfo contains the information needed to create a rectangle shape.
type RectangleInfo[S any] struct {

	// Filtering holds the collision-filtering metadata for the shape.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the shape.
	UserData S

	// Rectangle contains the rectangle information.
	//
	// It is specified in the local space of the object that the shape is
	// attached to.
	Rectangle shape2d.Rectangle
}

type objectShapeState[S any] struct {
	objectIndex    int32
	nextShapeIndex int32
	prevShapeIndex int32
	spatialID      query2d.TreeItemID
	filterRepresentation
	objectShapeRepresentation
	userData S
}

// objectShapesCanIntersect reports whether the specified two object shapes are
// allowed to be checked for intersection.
func objectShapesCanIntersect[S any](a, b *objectShapeState[S]) bool {
	if a.objectIndex >= b.objectIndex {
		return false // prevent self-intersection and repeated checks
	}
	return a.filterRepresentation.canInteractWith(&b.filterRepresentation)
}

type objectShapeRepresentation struct {
	lsBCircle shape2d.Circle
	wsBCircle shape2d.Circle

	lsTransform shape2d.Transform
	wsTransform shape2d.Transform

	kind       objectShapeKind
	points     []dprec.Vec2
	skinRadius float64
}

func (s *objectShapeRepresentation) update(parentTransform shape2d.Transform) {
	s.wsBCircle = shape2d.TransformedCircle(s.lsBCircle, parentTransform)

	s.wsTransform = shape2d.ChainedTransform(
		parentTransform,
		s.lsTransform,
	)
}

func (s *objectShapeRepresentation) gjkShape() gjk2d.Shape {
	return gjk2d.Shape{
		Position:   s.wsTransform.Translation,
		Rotation:   s.wsTransform.Rotation,
		Points:     s.points,
		SkinRadius: s.skinRadius,
	}
}

func (s *objectShapeRepresentation) toCircle() shape2d.Circle {
	return shape2d.Circle{
		Center: s.wsTransform.Translation,
		Radius: s.skinRadius,
	}
}

func (s *objectShapeRepresentation) toRectangle() shape2d.Rectangle {
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

type objectShapeKind uint32

const (
	objectShapeKindCircle objectShapeKind = iota
	objectShapeKindRectangle
	objectShapeKindCapsule
	objectShapeKindConvexHull
)
