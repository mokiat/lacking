package ecs

// MaxComponentCount returns the maximum number of components that can be
// created per scene.
//
// Unlike the max entity count, this number is not configurable.
const MaxComponentCount = 64

// ComponentSet represents a storage for components of the same type.
type ComponentSet[T any] interface {
	Set(entity Entity, value T)
	Unset(entity Entity)
	Ref(entity Entity) *T
	Mask() componentMask
}

type componentMask uint64
