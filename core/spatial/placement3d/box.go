package placement3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// BoxInfo contains the information needed to create a box shape.
type BoxInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Box contains the box information.
	Box shape3d.Box
}

type sceneBoxShape[S any] struct {
	sceneShape[S]
	boxSolver
}

func newBoxSolver(template shape3d.Box) boxSolver {
	return boxSolver{
		lsBox:            template,
		lsBoundingSphere: template.BoundingSphere(),
	}
}

type boxSolver struct {
	lsBox            shape3d.Box
	lsBoundingSphere shape3d.Sphere

	wsBox            shape3d.Box
	wsBoundingSphere shape3d.Sphere
}

func (s *boxSolver) update(transform shape3d.Transform) {
	s.wsBox = shape3d.TransformedBox(s.lsBox, transform)
	s.wsBoundingSphere = shape3d.TransformedSphere(s.lsBoundingSphere, transform)
}

func (s *boxSolver) boundingSphere() shape3d.Sphere {
	return s.wsBoundingSphere
}
