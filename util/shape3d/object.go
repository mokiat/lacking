package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

const invalidObjectIndex = uint32(0xFFFFFFFF)

const InvalidObjectID = ObjectID(invalidObjectIndex)

type ObjectID uint32

type ObjectInfo[T any] struct {
	Position opt.T[dprec.Vec3]
	Rotation opt.T[dprec.Quat]

	// SourceMask  opt.T[uint32]
	// TargetMask  opt.T[uint32]
	// RejectGroup uint32
	Static bool

	UserData T
}

type SizeTest = Object[struct{}] // TODO: DELETE ME

type Object[T any] struct {
	transform  Transform
	firstShape shapeRef
	flags      objectFlags
	userData   T
}

func (o *Object[T]) IsStatic() bool {
	return o.flags&objectFlagsStatic != 0
}

// func (o *Object[T]) TransformedBoundingSphere() Sphere {
// 	// TODO: Maybe cache this?
// 	return Sphere{
// 		Position: o.transform.Apply(o.boundingSphere.Position),
// 		Radius:   o.boundingSphere.Radius,
// 	}
// }

type objectFlags uint32

const (
	objectFlagsNone   objectFlags = 0
	objectFlagsStatic objectFlags = 1 << iota
)
