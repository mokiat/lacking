package physics

import (
	"github.com/mokiat/lacking/util/shape"
)

const (
	epsilon    = float64(0.0001)
	sqrEpsilon = epsilon * epsilon
)

var nextCollisionGroup = 1

func NewCollisionGroup() int {
	result := nextCollisionGroup
	nextCollisionGroup++
	return result
}

type CollisionShape = shape.Placement[shape.Shape]
