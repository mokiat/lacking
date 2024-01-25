package physics

import (
	"github.com/mokiat/lacking/game/physics/collision"
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
	reference    indexReference
	name         string
	collisionSet collision.Set
}
