package shape3d

import "github.com/mokiat/lacking/util/mem"

type SphereID struct {
	internalID mem.SparseID
}

func (i SphereID) IsNil() bool {
	return i == (SphereID{})
}

type SphereShape struct {
	id           mem.SparseID
	objectID     mem.SparseID
	nextSphereID mem.SparseID
	template     Sphere
}

func (s *SphereShape) Init(id mem.SparseID, template Sphere) {
	s.id = id
	s.template = template
}

func (s *SphereShape) BoundingSphere() Sphere {
	return s.template
}
