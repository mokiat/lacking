package shape3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/mem"
)

type MeshID struct {
	internalID mem.SparseID
}

func (i MeshID) IsNil() bool {
	return i == (MeshID{})
}

type MeshShape struct {
	id         mem.SparseID
	objectID   mem.SparseID
	nextMeshID mem.SparseID
	template   Mesh
}

func (s *MeshShape) Init(id mem.SparseID, template Mesh) {
	s.id = id
	s.template = template
}

func (s *MeshShape) BoundingSphere() Sphere {
	if true {
		panic("TODO")
	}
	return Sphere{ // FIXME: This is not correct,
		Position: dprec.ZeroVec3(),
		Radius:   1.0,
	}
}
