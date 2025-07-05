package shape3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/mem"
)

type BoxID struct {
	internalID mem.SparseID
}

func (i BoxID) IsNil() bool {
	return i == (BoxID{})
}

type BoxShape struct {
	id        mem.SparseID
	objectID  mem.SparseID
	nextBoxID mem.SparseID
	template  Box
}

func (s *BoxShape) Init(id mem.SparseID, template Box) {
	s.id = id
	s.template = template
}

func (s *BoxShape) BoundingSphere() Sphere {
	return Sphere{
		Position: s.template.Position,
		Radius: dprec.Sqrt(
			dprec.Sqr(s.template.HalfWidth) + dprec.Sqr(s.template.HalfHeight) + dprec.Sqr(s.template.HalfLength),
		),
	}
}
