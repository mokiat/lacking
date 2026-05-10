package ecs

import "github.com/mokiat/lacking/game/ecs/v6/internal"

// EditOperation represents a change to be applied to an entity's components.
//
// Instances of this type should not be created directly nor kept around, but
// instead should only be used within the scope of an EditEntity callback.
type EditOperation struct {
	stager        *internal.Stager
	commandBuffer *internal.Buffer
	stageRow      internal.Row
}

// EditOperationFunc is used to perform edits on an entity's components within
// an EditEntity callback.
type EditOperationFunc func(op *EditOperation)

// AddComponent adds a component of type T with the provided value to the entity
// being edited.
//
// The entity must not already have a component of the specified type, otherwise
// the call will lead to a panic.
func AddComponent[T any](op *EditOperation, compType ComponentType[T], value T) {
	anyColumn := op.stager.ComponentColumn(compType.id)
	column := anyColumn.(*internal.Column[T])
	column.SetValue(op.stageRow, value)

	internal.WriteToBuffer(op.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeAddComponent,
	})
	internal.WriteToBuffer(op.commandBuffer, internal.AddComponentCommand{
		TypeID: compType.id,
	})
}

// RemoveComponent removes the component of type T from the entity being edited.
//
// The entity must already have a component of the specified type, otherwise the
// call will lead to a panic.
func RemoveComponent[T any](op *EditOperation, compType ComponentType[T]) {
	internal.WriteToBuffer(op.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeRemoveComponent,
	})
	internal.WriteToBuffer(op.commandBuffer, internal.RemoveComponentCommand{
		TypeID: compType.id,
	})
}

// ReplaceComponent replaces the component of type T on the entity being edited
// with the provided value.
//
// The entity must already have a component of the specified type, otherwise the
// call will lead to a panic.
func ReplaceComponent[T any](op *EditOperation, compType ComponentType[T], value T) {
	anyColumn := op.stager.ComponentColumn(compType.id)
	column := anyColumn.(*internal.Column[T])
	column.SetValue(op.stageRow, value)

	internal.WriteToBuffer(op.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeReplaceComponent,
	})
	internal.WriteToBuffer(op.commandBuffer, internal.ReplaceComponentCommand{
		TypeID: compType.id,
	})
}

// ReadOperation represents a request to read components of an entity.
//
// Instances of this type should not be created directly nor kept around but
// instead should only be used within the scope of a ReadEntity callback.
type ReadOperation struct {
	mask internal.TypeMask
	row  internal.Row

	componentLookup  internal.TypeLookup
	componentColumns []internal.AnyColumn
}

// GetComponent retrieves the component of type T from the entity being read
// and returns a reference to it.
//
// If a component that the entity does not have is requested, nil is returned.
func GetComponent[T any](op *ReadOperation, compType ComponentType[T]) *T {
	if !op.mask.HasType(compType.id) {
		return nil
	}

	anyColumn := op.componentColumns[op.componentLookup[compType.id]]
	column := anyColumn.(*internal.Column[T])
	return column.RefValue(op.row)
}

// InjectComponent retrieves the component of type T from the entity being read
// and injects it into the provided target pointer.
//
// If you request a component that the entity does not have, the target will be
// set to nil.
func InjectComponent[T any](op *ReadOperation, compType ComponentType[T], target **T) {
	*target = GetComponent(op, compType)
}
