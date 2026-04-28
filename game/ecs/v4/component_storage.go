package ecs

import (
	"fmt"
	"reflect"

	"github.com/mokiat/gog"
)

// ComponentStorage represents a storage for components of a specific type.
//
// A large chunk of the memory that will be allocated and used by the ECS scene
// will be managed by component storages. It is possible to share storages
// between multiple scenes.
type ComponentStorage interface {
	reflectType() reflect.Type
}

// NewComponentStorage creates a new component storage for components of type T.
func NewComponentStorage[T any]() ComponentStorage {
	return &specificComponentStorage[T]{
		pendingValue: *new(T),
	}
}

type specificComponentStorage[T any] struct {
	// TODO: Slice of components that is handed out to archetype chunks.

	pendingValue T
}

func (s *specificComponentStorage[T]) reflectType() reflect.Type {
	panic("not implemented")
}

func getTypeIndex[T any](scene *Scene) typeIndex {
	tIndex, ok := scene.storageMapping[reflect.TypeFor[T]()]
	if !ok {
		panic(fmt.Errorf("type %T not registered with scene", gog.Zero[T]()))
	}
	return tIndex
}

func getStorageFromIndex[T any](scene *Scene, tIndex typeIndex) *specificComponentStorage[T] {
	ifaceStorage := scene.storages[tIndex]
	return ifaceStorage.(*specificComponentStorage[T])
}

func getStorage[T any](scene *Scene) *specificComponentStorage[T] {
	tIndex := getTypeIndex[T](scene)
	return getStorageFromIndex[T](scene, tIndex)
}
