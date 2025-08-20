package shape3d

import (
	"github.com/mokiat/gomath/dprec"
)

// BoxInfo contains the information needed to create a box shape.
type BoxInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Box contains the box information.
	Box Box
}

type sceneBoxShape[S any] struct {
	sceneShape[S]
	boxSolver
}

func newBoxSolver(template Box) boxSolver {
	return boxSolver{
		lsBox: template,
		lsBoundingSphere: Sphere{
			Position: template.Position,
			Radius: dprec.Sqrt(
				dprec.Sqr(template.HalfWidth) + dprec.Sqr(template.HalfHeight) + dprec.Sqr(template.HalfLength),
			),
		},
	}
}

type boxSolver struct {
	lsBox            Box
	lsBoundingSphere Sphere

	wsBox            Box
	wsBoundingSphere Sphere
}

func (s *boxSolver) update(transform Transform) {
	wsTransform := ChainedTransform(transform, Transform{
		Translation: s.lsBox.Position,
		Rotation:    s.lsBox.Rotation,
	})
	s.wsBox = Box{
		Position:   wsTransform.Translation,
		Rotation:   wsTransform.Rotation,
		HalfWidth:  s.lsBox.HalfWidth,
		HalfHeight: s.lsBox.HalfHeight,
		HalfLength: s.lsBox.HalfLength,
	}
	s.wsBoundingSphere = Sphere{
		Position: transform.Apply(s.lsBoundingSphere.Position),
		Radius:   s.lsBoundingSphere.Radius,
	}
}

func (s *boxSolver) boundingSphere() Sphere {
	return s.wsBoundingSphere
}
