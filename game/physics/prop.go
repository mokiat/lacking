package physics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape3d"
)

type PropInfo struct {
	Name             string
	Position         opt.T[dprec.Vec3]
	Rotation         opt.T[dprec.Quat]
	CollisionSpheres []shape3d.Sphere
	CollisionBoxes   []shape3d.Box
	CollisionMeshes  []shape3d.Mesh
}

type Prop struct {
	name string
}

func (p Prop) Name() string {
	return p.name
}

type propState struct {
	reference indexReference
	objectID  shape3d.ObjectID
	name      string
}
