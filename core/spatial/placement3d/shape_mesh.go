package placement3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

type meshShape[S any] struct {
	baseShape[S]
	meshShapeRepresentation
}

type meshShapeRepresentation struct {
	lsBSphere shape3d.Sphere
	wsBSphere shape3d.Sphere

	lsTriangles []shape3d.Triangle
	wsTriangles []shape3d.Triangle

	points [3]dprec.Vec3
}

func (s *meshShapeRepresentation) update(parentTransform shape3d.Transform) {
	s.wsBSphere = shape3d.TransformedSphere(s.lsBSphere, parentTransform)

	for i := range s.lsTriangles {
		s.wsTriangles[i] = shape3d.TransformedTriangle(s.lsTriangles[i], parentTransform)
	}
}

func (s *meshShapeRepresentation) boundingSphere() shape3d.Sphere {
	return s.wsBSphere
}

func (s *meshShapeRepresentation) gjkShapeCount() int {
	return len(s.wsTriangles)
}

func (s *meshShapeRepresentation) gjkShape(index int) gjk3d.Shape {
	triangle := &s.wsTriangles[index]
	points := s.points[:]
	points[0] = triangle.A
	points[1] = triangle.B
	points[2] = triangle.C
	return gjk3d.Shape{
		Position:   dprec.ZeroVec3(),
		Rotation:   shape3d.IdentityRotation(),
		Points:     points,
		SkinRadius: 0.0,
	}
}
