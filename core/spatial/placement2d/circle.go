package placement2d

import "github.com/mokiat/lacking/core/spatial/shape2d"

// CircleInfo contains the information needed to create a circle shape.
type CircleInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Circle contains the circle information.
	Circle shape2d.Circle
}

type sceneCircleShape[S any] struct {
	sceneShape[S]
	circleSolver
}

func newCircleSolver(template shape2d.Circle) circleSolver {
	return circleSolver{
		lsCircle: template,
		wsCircle: template,
	}
}

type circleSolver struct {
	lsCircle shape2d.Circle
	wsCircle shape2d.Circle
}

func (s *circleSolver) update(transform shape2d.Transform) {
	s.wsCircle = shape2d.TransformedCircle(s.lsCircle, transform)
}

func (s *circleSolver) boundingCircle() shape2d.Circle {
	return s.wsCircle
}
