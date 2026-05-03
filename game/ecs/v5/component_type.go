package ecs

import (
	"reflect"

	"github.com/mokiat/lacking/game/ecs/v5/internal"
)

const chunkSize = 64 // TODO: Experiment with different chunk sizes to find a good balance between memory usage and cache performance.

// BaseComponentType represents a component type in the ECS. It is used to
// identify component types and to manage component storage.
type BaseComponentType interface {
	// id() typeID
	reflectType() reflect.Type

	createChain() componentChain
	borrowChunk() uint32
	returnChunk(offset uint32)
	copyItem(dstChunk, dstOffset, srcChunk, srcOffset uint32)
}

// ComponentType represents a component type in the ECS. It is used to
// identify component types and to manage component storage.
type ComponentType[T any] struct {
	// TODO: NO LONGER use this type via pointer reference, but instead pass
	// by value. The mutability of the type is now handled inside the storage.

	id      internal.TypeID
	storage *internal.ComponentStorage[T]

	// TODO: REMOVE!
	tIndex typeID
	chunks [][chunkSize]T
}

func (t ComponentType[T]) baseStorage() internal.BaseComponentStorage {
	return t.storage
}

var _ BaseComponentType = (*ComponentType[any])(nil)

func newComponentType[T any](tIndex typeID) *ComponentType[T] {
	return &ComponentType[T]{
		tIndex: tIndex,
	}
}

// func (t *ComponentType[T]) id() typeID {
// 	return t.tIndex
// }

func (t *ComponentType[T]) reflectType() reflect.Type {
	return reflect.TypeFor[T]() // TODO: Cache value
}

func (t *ComponentType[T]) createChain() componentChain {
	// TODO: Fetch from a pool?
	return &specificComponentChain[T]{
		compType: t,
		chunks:   nil,
	}
}

func (t *ComponentType[T]) borrowChunk() uint32 {
	panic("not implemented")
}

func (t *ComponentType[T]) returnChunk(offset uint32) {
	panic("not implemented")
}

func (t *ComponentType[T]) copyItem(dstChunk, dstOffset, srcChunk, srcOffset uint32) {
	panic("not implemented")
}

func (t *ComponentType[T]) setValue(pos storagePosition, value T) {
	t.chunks[pos.chunkID][pos.chunkOffset] = value
}

func (t *ComponentType[T]) value(pos storagePosition) T {
	return t.chunks[pos.chunkID][pos.chunkOffset]
}

func (t *ComponentType[T]) refValue(pos storagePosition) *T {
	return &t.chunks[pos.chunkID][pos.chunkOffset]
}

type storagePosition struct {
	chunkID     uint32
	chunkOffset uint32
}

// typeID is a unique identifier for a component type.
type typeID uint32
