package shape3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/mem"
)

type MeshShape struct {
	nextMeshID mem.SparseID
	mesh       Mesh
}

func (s MeshShape) BoundingSphere() Sphere {
	return Sphere{ // FIXME: This is not correct,
		Position: dprec.ZeroVec3(),
		Radius:   1.0,
	}
}
