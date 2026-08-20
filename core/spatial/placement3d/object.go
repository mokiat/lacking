package placement3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// NilObjectID indicates an object that can never be part of the scene.
//
// It is also used to denote the absence of a source object in contacts that
// were produced by a query primitive rather than by a scene shape.
const NilObjectID = ObjectID(nilIndex)

// ObjectID is a reference to an object in the scene.
type ObjectID int32

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

	// UserData allows one to attach custom user data to an object.
	UserData O
}

type objectState[O any] struct {
	transform       shape3d.Transform
	firstShapeIndex int32
	lastShapeIndex  int32
	userData        O
}
