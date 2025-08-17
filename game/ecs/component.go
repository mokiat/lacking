package ecs

// MaxComponentCount returns the maximum number of components that can be
// created per scene.
//
// Unlike the max entity count, this number is not configurable.
const MaxComponentCount = 64

// ComponentSet represents a storage for components of the same type.
type ComponentSet[T any] interface {
	Set(entityID EntityID, value T)
	Unset(entityID EntityID)
	Ref(entityID EntityID) *T
	Mask() componentMask
}

type componentMask uint64
