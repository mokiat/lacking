package placement2d

import "github.com/mokiat/lacking/core/spatial/shape2d"

// RectangleInfo contains the information needed to create a rectangle shape.
type RectangleInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Rectangle contains the rectangle information.
	Rectangle shape2d.Rectangle
}

type sceneRectangleShape[S any] struct {
	sceneShape[S]
	rectangleSolver
}

func newRectangleSolver(template shape2d.Rectangle) rectangleSolver {
	return rectangleSolver{
		lsRectangle:      template,
		lsBoundingCircle: template.BoundingCircle(),
	}
}

type rectangleSolver struct {
	lsRectangle      shape2d.Rectangle
	lsBoundingCircle shape2d.Circle

	wsRectangle      shape2d.Rectangle
	wsBoundingCircle shape2d.Circle
}

func (s *rectangleSolver) update(transform shape2d.Transform) {
	s.wsRectangle = shape2d.TransformedRectangle(s.lsRectangle, transform)
	s.wsBoundingCircle = shape2d.TransformedCircle(s.lsBoundingCircle, transform)
}

func (s *rectangleSolver) boundingCircle() shape2d.Circle {
	return s.wsBoundingCircle
}
