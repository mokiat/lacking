package physics

var nextCollisionGroup = 1

func NewCollisionGroup() int {
	result := nextCollisionGroup
	nextCollisionGroup++
	return result
}

type sbCollisionPair struct {
	BodyRef indexReference
	PropRef indexReference
}

type dbCollisionPair struct {
	PrimaryRef   indexReference
	SecondaryRef indexReference
}
