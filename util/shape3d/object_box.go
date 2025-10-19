package shape3d

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
		lsBox:            template,
		lsBoundingSphere: template.BoundingSphere(),
	}
}

type boxSolver struct {
	lsBox            Box
	lsBoundingSphere Sphere

	wsBox            Box
	wsBoundingSphere Sphere
}

func (s *boxSolver) update(transform Transform) {
	s.wsBox = TransformedBox(s.lsBox, transform)
	s.wsBoundingSphere = TransformedSphere(s.lsBoundingSphere, transform)
}

func (s *boxSolver) boundingSphere() Sphere {
	return s.wsBoundingSphere
}
