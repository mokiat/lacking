package shape3d

import "github.com/mokiat/lacking/util/mem"

type CylinderID struct {
	internalID mem.SparseID
}

type CylinderShape struct {
	nextCylinderID mem.SparseID
	template       Cylinder
}

func (s *CylinderShape) BoundingSphere() Sphere {
	return Sphere{
		Position: s.template.Position,
		Radius:   s.template.Radius + s.template.HalfHeight, // FIXME: This is not correct.
	}
}
