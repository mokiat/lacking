package ecs

// EditOperation represents a change to be applied to an entity's components.
//
// Instances of this type should not be created directly nor kept around, but
// instead should only be used within the scope of an EditEntity callback.
type EditOperation struct {
	scene        *Scene
	mask         componentMask
	placeholders [MaxComponentTypes]storagePosition
}

// EditOperationFunc is used to perform edits on an entity's components within
// an EditEntity callback.
type EditOperationFunc func(op *EditOperation)

// AddComponent adds a component of type T with the provided value to the entity
// being edited.
func AddComponent[T any](op *EditOperation, compType *ComponentType[T], value T) {
	id := compType.id()
	if op.mask.containsType(id) {
		panic("entity already has component of this type")
	}
	op.mask.addType(id)

	// placeholder := op.placeholders[id]
	// compType.setValue(placeholder, value)
}

// RemoveComponent removes the component of type T from the entity being edited.
func RemoveComponent[T any](op *EditOperation, compType *ComponentType[T]) {
	id := compType.id()
	if !op.mask.containsType(id) {
		panic("entity does not have component of this type")
	}
	op.mask.removeType(id)
}

// ReadOperation represents a request to read components of an entity.
//
// Instances of this type should not be created directly nor kept around but
// instead should only be used within the scope of a ReadEntity callback.
type ReadOperation struct {
	scene           *Scene
	archetype       *componentArchetype
	archetypeOffset uint32
}

// // GetComponent retrieves the component of type T from the entity being read
// // and returns a reference to it.
// //
// // If you request a component that the entity does not have, nil is returned.
// func GetComponent[T any](op *ReadOperation) *T {
// 	tIndex := getTypeIndex[T](op.scene)

// 	mask := op.archetype.mask
// 	if !mask.containsType(tIndex) {
// 		return nil
// 	}

// 	anyChain := op.archetype.components[tIndex]
// 	chain := anyChain.(*specificComponentChain[T])
// 	return chain.getRef(op.archetypeOffset)
// }

// // InjectComponent retrieves the component of type T from the entity being read
// // and injects it into the provided target pointer.
// //
// // If you request a component that the entity does not have, the target will be
// // set to nil.
// func InjectComponent[T any](op *ReadOperation, target **T) {
// 	*target = GetComponent[T](op)
// }
