package physics

var nextCollisionGroup = 1

func NewCollisionGroup() int {
	result := nextCollisionGroup
	nextCollisionGroup++
	return result
}
