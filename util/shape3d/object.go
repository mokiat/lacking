package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// InvalidObjectID indicates an object that can never be part of the scene.
const InvalidObjectID = ObjectID(invalidObjectIndex)

// ObjectID is a reference to an object in the scene.
type ObjectID uint32

// ObjectInfo contains the information needed to create an object in a scene.
type ObjectInfo[O any] struct {

	// Position optionally specifies a position where the object should be
	// placed.
	//
	// Defaults to the origin.
	Position opt.T[dprec.Vec3]

	// Rotation optionally specifies a rotation of the object.
	//
	// Defaults to the identity rotation.
	Rotation opt.T[dprec.Quat]

	// Static marks the object as static. Static objects are not checked for
	// intersections with other static objects.
	Static bool

	// UserData allows one to attach custom user data to an object.
	UserData O
}

type sceneObject[O any] struct {
	transform  Transform
	firstShape shapeRef
	flags      objectFlags
	userData   O
}

func (o *sceneObject[T]) isStatic() bool {
	return o.flags&objectFlagsStatic != 0
}

const invalidObjectIndex = uint32(0xFFFFFFFF)

type objectFlags uint32

const (
	objectFlagsNone   objectFlags = 0
	objectFlagsStatic objectFlags = 1 << iota
)
