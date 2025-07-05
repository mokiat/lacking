package shape3d

type SphereShape struct {
	template Sphere
	solver   sphereSolver
}

func (s *SphereShape) Init(template Sphere, transform Transform) {
	s.template = template
	s.solver.Update(template, transform)
}

func (s *SphereShape) SetTransform(transform Transform) {
	s.solver.Update(s.template, transform)
}

func (s *SphereShape) BoundingSphere() Sphere {
	return s.template
}

type sphereSolver struct {
	wsSphere Sphere
}

func (s *sphereSolver) Update(template Sphere, transform Transform) {
	s.wsSphere = Sphere{
		Position: transform.Apply(template.Position),
		Radius:   template.Radius,
	}
}
