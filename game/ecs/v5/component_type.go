package ecs

import "reflect"

// BaseComponentType represents a component type in the ECS. It is used to
// identify component types and to manage component storage.
type BaseComponentType interface {
	typeIndex() typeIndex
	reflectType() reflect.Type

	borrowChunk() uint32
	returnChunk(offset uint32)
	copyItem(dstChunk, dstOffset, srcChunk, srcOffset uint32)
}

// ComponentType represents a component type in the ECS. It is used to
// identify component types and to manage component storage.
type ComponentType[T any] struct {
	tIndex typeIndex
}

var _ BaseComponentType = (*ComponentType[any])(nil)

func newComponentType[T any](tIndex typeIndex) *ComponentType[T] {
	return &ComponentType[T]{
		tIndex: tIndex,
	}
}
