package shape3d

// SphereInfo contains the information needed to create a sphere shape.
type SphereInfo struct {

	// ShapeInfo contains general shape information.
	ShapeInfo

	// Sphere contains the sphere information.
	Sphere Sphere
}

type sceneSphereShape struct {
	sceneShape
	sphereSolver
}

func newSphereSolver(template Sphere) sphereSolver {
	return sphereSolver{
		lsSphere: template,
		wsSphere: template,
	}
}

type sphereSolver struct {
	lsSphere Sphere
	wsSphere Sphere
}

func (s *sphereSolver) update(transform Transform) {
	s.wsSphere = Sphere{
		Position: transform.Apply(s.lsSphere.Position),
		Radius:   s.lsSphere.Radius,
	}
}

func (s *sphereSolver) boundingSphere() Sphere {
	return s.wsSphere
}
