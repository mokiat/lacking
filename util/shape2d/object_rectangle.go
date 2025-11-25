package shape2d

// RectangleInfo contains the information needed to create a rectangle shape.
type RectangleInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Rectangle contains the rectangle information.
	Rectangle Rectangle
}

type sceneRectangleShape[S any] struct {
	sceneShape[S]
	rectangleSolver
}

func newRectangleSolver(template Rectangle) rectangleSolver {
	return rectangleSolver{
		lsRectangle:      template,
		lsBoundingCircle: template.BoundingCircle(),
	}
}

type rectangleSolver struct {
	lsRectangle      Rectangle
	lsBoundingCircle Circle

	wsRectangle      Rectangle
	wsBoundingCircle Circle
}

func (s *rectangleSolver) update(transform Transform) {
	s.wsRectangle = TransformedRectangle(s.lsRectangle, transform)
	s.wsBoundingCircle = TransformedCircle(s.lsBoundingCircle, transform)
}

func (s *rectangleSolver) boundingCircle() Circle {
	return s.wsBoundingCircle
}
