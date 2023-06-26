package physics

import (
	"github.com/mokiat/lacking/game/physics/collision"
)

type PropInfo struct {
	CollisionSet collision.Set
}

type Prop struct {
	collisionSet collision.Set
}
