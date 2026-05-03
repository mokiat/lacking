package ecs

import "github.com/mokiat/lacking/game/ecs/v5/internal"

// EditOperation represents a change to be applied to an entity's components.
//
// Instances of this type should not be created directly nor kept around, but
// instead should only be used within the scope of an EditEntity callback.
type EditOperation struct {
	scene        *Scene
	mask         internal.TypeMask
	placeholders [MaxComponentTypes]storagePosition
}

// EditOperationFunc is used to perform edits on an entity's components within
// an EditEntity callback.
type EditOperationFunc func(op *EditOperation)

// AddComponent adds a component of type T with the provided value to the entity
// being edited.
func AddComponent[T any](op *EditOperation, compType ComponentType[T], value T) {
	if op.mask.HasType(compType.id) {
		panic("entity already has component of this type")
	}
	op.mask.AddType(compType.id)

	// FIXME!
	// placeholder := op.placeholders[compType.id]
	// compType.setValue(placeholder, value)
}

// RemoveComponent removes the component of type T from the entity being edited.
func RemoveComponent[T any](op *EditOperation, compType ComponentType[T]) {
	if !op.mask.HasType(compType.id) {
		panic("entity does not have component of this type")
	}
	op.mask.RemoveType(compType.id)
}

// ReadOperation represents a request to read components of an entity.
//
// Instances of this type should not be created directly nor kept around but
// instead should only be used within the scope of a ReadEntity callback.
type ReadOperation struct {
	scene        *Scene
	archetype    *internal.Archetype
	archetypeRow internal.ArchetypeRow
}

// GetComponent retrieves the component of type T from the entity being read
// and returns a reference to it.
//
// If a component that the entity does not have is requested, nil is returned.
func GetComponent[T any](op *ReadOperation, compType ComponentType[T]) *T {
	mask := op.archetype.TypeMask()
	if !mask.HasType(compType.id) {
		return nil
	}

	// TODO: Get Column (virtual representation)
	chain := getChain(op.archetype, compType)
	return chain.getRef(op.archetypeOffset)
}

// InjectComponent retrieves the component of type T from the entity being read
// and injects it into the provided target pointer.
//
// If you request a component that the entity does not have, the target will be
// set to nil.
func InjectComponent[T any](op *ReadOperation, compType ComponentType[T], target **T) {
	*target = GetComponent(op, compType)
}
