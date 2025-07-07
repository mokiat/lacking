package shape3d

import (
	"github.com/mokiat/gomath/dprec"
)

type BoxInfo struct {
	ShapeInfo
	Box Box
}

type BoxShape struct {
	Shape
	boxSolver
}

func newBoxSolver(template Box) boxSolver {
	return boxSolver{
		template: template,
		boundingSphere: Sphere{
			Position: template.Position,
			Radius: dprec.Sqrt(
				dprec.Sqr(template.HalfWidth) + dprec.Sqr(template.HalfHeight) + dprec.Sqr(template.HalfLength),
			),
		},
	}
}

type boxSolver struct {
	template       Box
	boundingSphere Sphere

	wsBox            Box
	wsBoundingSphere Sphere
}

func (s *boxSolver) Update(transform Transform) {
	wsTransform := ChainedTransform(transform, Transform{
		Translation: s.template.Position,
		Rotation:    s.template.Rotation,
	})
	s.wsBox = Box{
		Position:   wsTransform.Translation,
		Rotation:   wsTransform.Rotation,
		HalfWidth:  s.template.HalfWidth,
		HalfHeight: s.template.HalfHeight,
		HalfLength: s.template.HalfLength,
	}
	s.wsBoundingSphere = Sphere{
		Position: transform.Apply(s.boundingSphere.Position),
		Radius:   s.boundingSphere.Radius,
	}
}

func (s *boxSolver) BoundingSphere() Sphere {
	return s.wsBoundingSphere
}
