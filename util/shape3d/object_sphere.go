package shape3d

type SphereInfo struct {
	ShapeInfo
	Sphere Sphere
}

type SphereShape struct {
	Shape
	sphereSolver
}

func newSphereSolver(template Sphere) sphereSolver {
	return sphereSolver{
		template: template,
		wsSphere: template,
	}
}

type sphereSolver struct {
	template Sphere
	wsSphere Sphere
}

func (s *sphereSolver) Update(transform Transform) {
	s.wsSphere = Sphere{
		Position: transform.Apply(s.template.Position),
		Radius:   s.template.Radius,
	}
}

func (s *sphereSolver) BoundingSphere() Sphere {
	return s.wsSphere
}
