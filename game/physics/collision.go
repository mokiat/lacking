package physics

import (
	"github.com/mokiat/lacking/util/shape"
)

var nextCollisionGroup = 1

func NewCollisionGroup() int {
	result := nextCollisionGroup
	nextCollisionGroup++
	return result
}

type CollisionShape = shape.Placement[shape.Shape]
