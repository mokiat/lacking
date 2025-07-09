package physics

import (
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/util/shape3d"
)

type PropInfo struct {
	Name         string
	CollisionSet collision.Set
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
