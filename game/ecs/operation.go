package ecs

import "github.com/mokiat/lacking/game/ecs/internal"

// EditOperation is the write handle passed to [Scene.EditEntity] and
// [Scene.CreateEntity] callbacks. Use [SetComponent] and [UnsetComponent]
// to stage component changes.
//
// Do not create instances directly or retain the pointer beyond the
// callback scope.
type EditOperation struct {
	stager        *internal.Stager
	commandBuffer *internal.Buffer
	stageRow      internal.Row
}

// EditOperationFunc is the callback signature accepted by
// [Scene.EditEntity] and [Scene.CreateEntity].
type EditOperationFunc func(op *EditOperation)

// SetComponent stages the addition of a component of type T with the
// given value to the entity being edited. If the entity already has a component
// of type T, it is replaced with the new value.
func SetComponent[T any](op *EditOperation, compType ComponentType[T], value T) {
	columnID := op.stager.ComponentColumnID(compType.id)
	column := compType.storage.Column(columnID)
	column.SetValue(op.stageRow, value)

	internal.WriteToBuffer(op.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeSetComponent,
	})
	internal.WriteToBuffer(op.commandBuffer, internal.SetComponentCommand{
		TypeID: compType.id,
	})
}

// UnsetComponent stages the removal of the component of type T from
// the entity being edited. If the entity does not have a component of type T,
// this is a no-op.
func UnsetComponent[T any](op *EditOperation, compType ComponentType[T]) {
	internal.WriteToBuffer(op.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeUnsetComponent,
	})
	internal.WriteToBuffer(op.commandBuffer, internal.UnsetComponentCommand{
		TypeID: compType.id,
	})
}

// ReadOperation is the read handle passed to [Scene.ReadEntity] and
// [Scene.QueryEntities] callbacks. Use [GetComponent] or
// [InjectComponent] to retrieve component values.
//
// Do not create instances directly or retain the pointer beyond the
// callback scope.
type ReadOperation struct {
	mask internal.TypeMask
	row  internal.Row

	componentLookup    internal.TypeLookup
	componentColumnIDs []internal.ColumnID
}

// GetComponent returns a pointer to the component of type T for the
// entity currently being read, or nil if the entity does not have the
// component.
func GetComponent[T any](op *ReadOperation, compType ComponentType[T]) *T {
	if !op.mask.HasType(compType.id) {
		return nil
	}

	columnID := op.componentColumnIDs[op.componentLookup[compType.id]]
	column := compType.storage.Column(columnID)
	return column.RefValue(op.row)
}

// InjectComponent sets *target to the component of type T for the
// entity currently being read, or nil if the entity does not have the
// component. It is a convenience wrapper around [GetComponent].
func InjectComponent[T any](op *ReadOperation, compType ComponentType[T], target **T) {
	*target = GetComponent(op, compType)
}
