package shape2d

// CircleInfo contains the information needed to create a circle shape.
type CircleInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Circle contains the circle information.
	Circle Circle
}

type sceneCircleShape[S any] struct {
	sceneShape[S]
	circleSolver
}

type circleSolver struct {
	lsCircle Circle
	wsCircle Circle
}

func (s *circleSolver) update(transform Transform) {
	s.wsCircle = TransformedCircle(s.lsCircle, transform)
}

func (s *circleSolver) boundingCircle() Circle {
	return s.wsCircle
}
