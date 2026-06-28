package placement3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// SphereInfo contains the information needed to create a sphere shape.
type SphereInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Sphere contains the sphere information.
	Sphere shape3d.Sphere
}

type sceneSphereShape[S any] struct {
	sceneShape[S]
	sphereSolver
}

func newSphereSolver(template shape3d.Sphere) sphereSolver {
	return sphereSolver{
		lsSphere: template,
		wsSphere: template,
	}
}

type sphereSolver struct {
	lsSphere shape3d.Sphere
	wsSphere shape3d.Sphere
}

func (s *sphereSolver) update(transform shape3d.Transform) {
	s.wsSphere = shape3d.TransformedSphere(s.lsSphere, transform)
}

func (s *sphereSolver) boundingSphere() shape3d.Sphere {
	return s.wsSphere
}
