package shape2d

import "slices"

// PolygonInfo contains the information needed to create a polygon shape.
type PolygonInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Polygon contains the polygon information.
	Polygon Polygon
}

type scenePolygonShape[S any] struct {
	sceneShape[S]
	polygonSolver
}

func newPolygonSolver(template Polygon) polygonSolver {
	bc := template.BoundingCircle()
	return polygonSolver{
		lsPolygon:        template,
		lsBoundingCircle: bc,

		wsPolygon:        NewPolygon(slices.Clone(template.Segments)),
		wsBoundingCircle: bc,
	}
}

type polygonSolver struct {
	lsPolygon        Polygon
	lsBoundingCircle Circle

	wsPolygon        Polygon
	wsBoundingCircle Circle
}

func (s *polygonSolver) update(transform Transform) {
	for i := range s.wsPolygon.Segments {
		s.wsPolygon.Segments[i] = TransformedSegment(
			s.lsPolygon.Segments[i],
			transform,
		)
	}
	s.wsBoundingCircle = TransformedCircle(s.lsBoundingCircle, transform)
}

func (s *polygonSolver) boundingCircle() Circle {
	return s.wsBoundingCircle
}
