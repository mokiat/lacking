package ecs

// ComponentTypeID is an identifier for a component type.
// Numbers should be in the range [0..63]. That is, the
// ECS framework supports at most 64 component types at
// the moment.
type ComponentTypeID uint8

func (i ComponentTypeID) mask() uint64 {
	return 1 << i
}
