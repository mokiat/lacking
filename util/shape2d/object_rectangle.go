package shape2d

import "github.com/mokiat/gomath/dprec"

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
		lsRectangle: template,
		lsBoundingCircle: Circle{
			Position: template.Position,
			Radius: dprec.Sqrt(
				dprec.Sqr(template.HalfWidth) + dprec.Sqr(template.HalfHeight),
			),
		},
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
	s.wsBoundingCircle = Circle{
		Position: transform.Apply(s.lsBoundingCircle.Position),
		Radius:   s.lsBoundingCircle.Radius,
	}
}

func (s *rectangleSolver) boundingCircle() Circle {
	return s.wsBoundingCircle
}
