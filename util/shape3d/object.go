package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/mem"
	"github.com/mokiat/lacking/util/spatial"
)

func NilObjectID() ObjectID {
	return ObjectID{}
}

type ObjectInfo[T any] struct {
	Position opt.T[dprec.Vec3]
	Rotation opt.T[dprec.Quat]

	Static      bool
	SourceGroup uint64
	SourceMask  opt.T[uint32]
	TargetMask  opt.T[uint32]

	UserData T

	// Insert specifies whether to immediatelly insert the object
	// into the scene. Adding shapes to an inserted object can be slower
	// so in certain situations it might make sense to defer that operation.
	Insert opt.T[bool]
}

type ObjectID struct {
	internalID mem.SparseID
}

func (i ObjectID) IsNil() bool {
	return i == (ObjectID{})
}

type SizeTest = Object[struct{}] // TODO: DELETE ME

type Object[T any] struct {
	id mem.SparseID

	transform Transform

	static      bool
	sourceGroup uint64
	sourceMask  uint32
	targetMask  uint32

	firstSphereID mem.SparseID
	firstBoxID    mem.SparseID
	firstMeshID   mem.SparseID

	spatialID      opt.T[spatial.DynamicOctreeItemID]
	boundingSphere Sphere

	userData T
}

func (o *Object[T]) ObjectID() ObjectID {
	return ObjectID{
		internalID: o.id,
	}
}

func (o *Object[T]) TransformedBoundingSphere() Sphere {
	// TODO: Maybe cache this?
	return Sphere{
		Position: o.transform.Apply(o.boundingSphere.Position),
		Radius:   o.boundingSphere.Radius,
	}
}
