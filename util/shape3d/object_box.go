package shape3d

import "github.com/mokiat/lacking/util/mem"

type BoxShape struct {
	nextBoxID mem.SparseID
	template  Box
}

func (s *BoxShape) BoundingSphere() Sphere {
	return Sphere{
		Position: s.template.Position,
		Radius:   s.template.HalfWidth + s.template.HalfHeight + s.template.HalfLength, // FIXME: This is not correct.
	}
}
