package ecs

import (
	"fmt"
	"iter"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/ecs/v6/internal"
	"github.com/mokiat/lacking/util/observer"
)

// NewScene creates a [Scene] backed by the component types registered
// in scope. A scene owns its own archetype storage and entity table and
// does not share data with other scenes.
func NewScene(scope *Scope) *Scene {
	scope.markInUse()

	const initialReadOperations = 2
	readOperations := ds.PreallocatedStack[*ReadOperation](initialReadOperations)
	for range initialReadOperations {
		readOperations.Push(new(ReadOperation))
	}

	const initialEditOperations = 1
	editOperations := ds.PreallocatedStack[*EditOperation](initialEditOperations)
	for range initialEditOperations {
		editOperations.Push(new(EditOperation))
	}

	return &Scene{
		registry: scope.registry,

		enterSubscriptions: observer.NewSubscriptionSet[ConditionalCallback](),
		exitSubscriptions:  observer.NewSubscriptionSet[ConditionalCallback](),

		freeEntityIndices: ds.EmptyStack[uint32](),
		entities:          nil,

		archetypePool: ds.EmptyStack[*internal.Archetype](),
		archetypes:    make(map[internal.TypeMask]*internal.Archetype),

		commandBuffer: internal.NewBuffer(1024), // 1KB initial capacity
		stager:        internal.NewStager(scope.registry),

		readOperations: readOperations,
		editOperations: editOperations,
	}
}

// Scene is the central ECS container. It stores entities and their
// components in archetype-grouped tables and provides methods for
// creating, deleting, reading, editing, and querying entities, as well
// as subscribing to structural change events.
type Scene struct {
	registry *internal.Registry

	enterSubscriptions *observer.SubscriptionSet[ConditionalCallback]
	exitSubscriptions  *observer.SubscriptionSet[ConditionalCallback]

	freeEntityIndices *ds.Stack[uint32]
	entities          []internal.Entity

	archetypePool *ds.Stack[*internal.Archetype]
	archetypes    map[internal.TypeMask]*internal.Archetype

	commandBuffer *internal.Buffer
	stager        *internal.Stager

	readOperations    *ds.Stack[*ReadOperation]
	editOperations    *ds.Stack[*EditOperation]
	queryDepth        uint32
	inOperationBlock  bool
	inQueueProcessing bool
}

// Destroy releases all resources held by the scene. The scene must not
// be used after this call.
func (s *Scene) Destroy() {
	s.enterSubscriptions.Clear()
	s.exitSubscriptions.Clear()
	for _, archetype := range s.archetypes {
		archetype.Destroy()
	}
	s.stager.Destroy()
}

// SubscribeEnter registers a callback that fires whenever an entity
// transitions into satisfying condition. The callback receives the ID
// of the entity that triggered the transition. Call Delete on the
// returned [EntitySubscription] to unsubscribe.
func (s *Scene) SubscribeEnter(condition Condition, callback EntityCallback) *EntitySubscription {
	return s.enterSubscriptions.Subscribe(ConditionalCallback{
		condition: condition,
		callback:  callback,
	})
}

// SubscribeExit registers a callback that fires whenever an entity
// transitions out of satisfying condition. The callback receives the ID
// of the entity that triggered the transition. Call Delete on the
// returned [EntitySubscription] to unsubscribe.
func (s *Scene) SubscribeExit(condition Condition, callback EntityCallback) *EntitySubscription {
	return s.exitSubscriptions.Subscribe(ConditionalCallback{
		condition: condition,
		callback:  callback,
	})
}

// CreateEntity allocates a new entity and returns its [ID]. If fn is
// not nil, fn is called with an [EditOperation] so that initial
// components can be added before the entity is committed to the scene.
//
// CreateEntity may be called during a query; the creation is deferred
// until the query completes.
func (s *Scene) CreateEntity(fn EditOperationFunc) ID {
	s.verifyOutsideOperation()

	index := s.allocateEntityIndex()
	desc := &s.entities[index]
	desc.Revive(index)

	stageRow := s.stager.Grow()

	internal.WriteToBuffer(s.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeCreateEntity,
	})
	internal.WriteToBuffer(s.commandBuffer, internal.CreateEntityCommand{
		EntityID: desc.ID(),
		StageRow: stageRow,
	})

	if fn != nil {
		editOperation := s.allocateEditOperation()
		defer s.releaseEditOperation(editOperation)

		*editOperation = EditOperation{
			stager:        s.stager,
			commandBuffer: s.commandBuffer,
			stageRow:      stageRow,
		}
		s.inOperationBlock = true
		fn(editOperation)
		s.inOperationBlock = false
	}

	internal.WriteToBuffer(s.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeEndOfSequence,
	})

	s.processQueue()

	return fromInternalID(desc.ID())
}

// DeleteEntity removes the entity and all its components from the
// scene. Any component pointers previously obtained for this entity
// must not be used after this call.
//
// DeleteEntity may be called during a query; the deletion is deferred
// until the query completes.
func (s *Scene) DeleteEntity(id ID) {
	s.verifyOutsideOperation()

	internal.WriteToBuffer(s.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeDeleteEntity,
	})
	internal.WriteToBuffer(s.commandBuffer, internal.DeleteEntityCommand{
		EntityID: internal.NewID(id.index, id.revision),
	})

	s.processQueue()
}

// HasEntity reports whether id refers to a live entity in the scene.
func (s *Scene) HasEntity(id ID) bool {
	s.verifyOutsideOperation()

	_, ok := s.getEntityDescriptor(id)
	return ok
}

// CheckEntity reports whether the entity identified by id satisfies
// condition. Returns false for invalid or deleted IDs.
func (s *Scene) CheckEntity(id ID, condition Condition) bool {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		return false
	}
	archetype := desc.Archetype()

	return condition.isSatisfiedBy(archetype.TypeMask())
}

// ReadEntity calls fn with a [ReadOperation] scoped to the entity
// identified by id. Use [GetComponent] or [InjectComponent] inside fn
// to retrieve component values.
//
// Panics if the entity does not exist.
func (s *Scene) ReadEntity(id ID, fn func(*ReadOperation)) {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic("entity does not exist")
	}

	archetype := desc.Archetype()
	mask := archetype.TypeMask()
	row := desc.ArchetypeRow()
	columnIDs, lookup := archetype.ComponentColumnIDs()

	readOperation := s.allocateReadOperation()
	defer s.releaseReadOperation(readOperation)
	readOperation.mask = mask
	readOperation.componentLookup = lookup
	readOperation.componentColumnIDs = columnIDs
	readOperation.row = row

	s.inOperationBlock = true
	fn(readOperation)
	s.inOperationBlock = false
}

// EditEntity calls fn with an [EditOperation] for the entity identified
// by id. Use [AddComponent], [RemoveComponent], and [ReplaceComponent]
// inside fn to stage structural or value changes.
//
// Panics if a component is added that the entity already has, or one is
// removed or replaced that the entity does not have, as determined by
// the virtual component state after each prior operation in the same
// edit. When multiple operations target the same component type, only
// the last one takes effect.
//
// EditEntity may be called during a query; the edit is deferred until
// the query completes.
func (s *Scene) EditEntity(id ID, fn EditOperationFunc) {
	s.verifyOutsideOperation()

	stageRow := s.stager.Grow()

	internal.WriteToBuffer(s.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeEditEntity,
	})
	internal.WriteToBuffer(s.commandBuffer, internal.EditEntityCommand{
		EntityID: internal.NewID(id.index, id.revision),
		StageRow: stageRow,
	})

	editOperation := s.allocateEditOperation()
	defer s.releaseEditOperation(editOperation)

	*editOperation = EditOperation{
		stager:        s.stager,
		commandBuffer: s.commandBuffer,
		stageRow:      stageRow,
	}
	s.inOperationBlock = true
	fn(editOperation)
	s.inOperationBlock = false

	internal.WriteToBuffer(s.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeEndOfSequence,
	})

	s.processQueue()
}

// QueryEntities calls yield for every entity that satisfies condition,
// passing a [ReadOperation] through which its components can be read.
// Returning false from yield stops the iteration early.
//
// Mutations made during the query via [Scene.EditEntity],
// [Scene.CreateEntity], or [Scene.DeleteEntity] are deferred and
// applied after iteration completes.
func (s *Scene) QueryEntities(condition Condition, yield func(ID, *ReadOperation) bool) {
	s.verifyOutsideOperation()

	s.queryDepth++

	readOperation := s.allocateReadOperation()
	defer s.releaseReadOperation(readOperation)

iteration:
	for mask, archetype := range s.archetypes {
		if !condition.isSatisfiedBy(mask) {
			continue
		}

		columnIDs, lookup := archetype.ComponentColumnIDs()

		readOperation.mask = mask
		readOperation.componentLookup = lookup
		readOperation.componentColumnIDs = columnIDs

		for row := internal.Row(0); uint32(row) < archetype.Size(); row++ {
			readOperation.row = row
			intID := archetype.IDColumn().Value(row)
			if !yield(fromInternalID(intID), readOperation) {
				break iteration
			}
		}
	}

	s.queryDepth--
	s.processQueue()
}

// QueryEntitiesIter returns a range iterator over entities that satisfy
// condition. It is equivalent to [Scene.QueryEntities] and carries the
// same deferred-mutation semantics.
func (s *Scene) QueryEntitiesIter(condition Condition) iter.Seq2[ID, *ReadOperation] {
	return func(yield func(ID, *ReadOperation) bool) {
		s.QueryEntities(condition, yield)
	}
}

func (s *Scene) verifyOutsideOperation() {
	if s.inOperationBlock {
		panic("cannot call this method from inside an operation block")
	}
}

func (s *Scene) allocateEntityIndex() uint32 {
	var index uint32
	if s.freeEntityIndices.IsEmpty() {
		index = uint32(len(s.entities))
		s.entities = append(s.entities, internal.Entity{})
	} else {
		index = s.freeEntityIndices.Pop()
	}
	return index
}

func (s *Scene) releaseEntityIndex(index uint32) {
	s.freeEntityIndices.Push(index)
}

func (s *Scene) getEntityDescriptor(id ID) (*internal.Entity, bool) {
	if id == NilID {
		return nil, false
	}
	desc := &s.entities[id.index]
	if !desc.HasRevision(id.revision) {
		return nil, false
	}
	return desc, true
}

func (s *Scene) borrowArchetypeRow(mask internal.TypeMask) (*internal.Archetype, internal.Row) {
	archetype, ok := s.archetypes[mask]
	if !ok {
		archetype = s.allocateArchetype(mask)
	}

	row := archetype.Grow()
	return archetype, row
}

func (s *Scene) releaseArchetypeRow(archetype *internal.Archetype, row internal.Row) {
	lastRow := archetype.LastRow()
	if row != lastRow {
		lastRowID := archetype.IDColumn().Value(lastRow)
		if lastRowID != internal.NilID {
			lastRowDesc := &s.entities[lastRowID.Index]
			lastRowDesc.Assign(archetype, row)
		}
		archetype.CopyRow(row, lastRow)
	}

	archetype.Shrink()

	if archetype.IsEmpty() {
		s.releaseArchetype(archetype)
	}
}

func (s *Scene) allocateArchetype(mask internal.TypeMask) *internal.Archetype {
	var result *internal.Archetype
	if s.archetypePool.IsEmpty() {
		result = internal.NewArchetype(s.registry)
	} else {
		result = s.archetypePool.Pop()
	}

	result.Revive(mask)

	s.archetypes[mask] = result

	return result
}

func (s *Scene) releaseArchetype(archetype *internal.Archetype) {
	mask := archetype.TypeMask()
	delete(s.archetypes, mask)

	archetype.Destroy()

	s.archetypePool.Push(archetype)
}

func (s *Scene) allocateReadOperation() *ReadOperation {
	if s.readOperations.IsEmpty() {
		return &ReadOperation{}
	}
	return s.readOperations.Pop()
}

func (s *Scene) releaseReadOperation(op *ReadOperation) {
	s.readOperations.Push(op)
}

func (s *Scene) allocateEditOperation() *EditOperation {
	if s.editOperations.IsEmpty() {
		return &EditOperation{}
	}
	return s.editOperations.Pop()
}

func (s *Scene) releaseEditOperation(op *EditOperation) {
	s.editOperations.Push(op)
}

func (s *Scene) processQueue() {
	if s.inQueueProcessing || s.queryDepth > 0 {
		return
	}
	s.inQueueProcessing = true
	defer func() {
		s.inQueueProcessing = false
	}()

	for s.commandBuffer.HasMoreData() {
		header := internal.ReadFromBuffer[internal.CommandHeader](s.commandBuffer)
		switch header.CommandType {

		case internal.CommandTypeCreateEntity:
			cmd := internal.ReadFromBuffer[internal.CreateEntityCommand](s.commandBuffer)
			s.processCreateEntityCommand(cmd)

		case internal.CommandTypeEditEntity:
			cmd := internal.ReadFromBuffer[internal.EditEntityCommand](s.commandBuffer)
			s.processEditEntityCommand(cmd)

		case internal.CommandTypeDeleteEntity:
			cmd := internal.ReadFromBuffer[internal.DeleteEntityCommand](s.commandBuffer)
			s.processDeleteEntityCommand(cmd)

		default:
			panic(fmt.Errorf("unexpected command type %v in scene command buffer", header.CommandType))
		}
	}

	s.commandBuffer.Reset()
	s.stager.Clear()
}

func (s *Scene) processCreateEntityCommand(cmd internal.CreateEntityCommand) {
	id := fromInternalID(cmd.EntityID)

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic(fmt.Errorf("entity with ID %v does not exist", id))
	}

	oldMask := opt.Unspecified[internal.TypeMask]() // starting from limbo
	newMask, changes := s.processComponentCommands(internal.EmptyTypeMask())

	archetype, row := s.borrowArchetypeRow(newMask)
	changes.EachType(func(id internal.TypeID) {
		storage := s.registry.Storage(id)
		newColumnID := archetype.ComponentColumnID(id)
		stageColumnID := s.stager.ComponentColumnID(id)
		storage.CopyCell(newColumnID, row, stageColumnID, cmd.StageRow)
	})

	desc.Assign(archetype, row)

	s.dispatchEnterEvent(id, oldMask, newMask)
}

func (s *Scene) processEditEntityCommand(cmd internal.EditEntityCommand) {
	id := fromInternalID(cmd.EntityID)

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic(fmt.Errorf("entity with ID %v does not exist", id))
	}

	oldMask := desc.Archetype().TypeMask()
	newMask, changes := s.processComponentCommands(oldMask)

	if oldMask != newMask {
		s.dispatchExitEvent(id, oldMask, newMask)

		oldArchetype := desc.Archetype()
		oldRow := desc.ArchetypeRow()

		newArchetype, newRow := s.borrowArchetypeRow(newMask)
		newMask.EachType(func(id internal.TypeID) {
			storage := s.registry.Storage(id)
			newColumnID := newArchetype.ComponentColumnID(id)
			if changes.HasType(id) {
				stageColumnID := s.stager.ComponentColumnID(id)
				storage.CopyCell(newColumnID, newRow, stageColumnID, cmd.StageRow)
			} else {
				oldColumnID := oldArchetype.ComponentColumnID(id)
				storage.CopyCell(newColumnID, newRow, oldColumnID, oldRow)
			}
		})
		s.releaseArchetypeRow(oldArchetype, oldRow)

		desc.Assign(newArchetype, newRow)

		s.dispatchEnterEvent(id, opt.V(oldMask), newMask)
	} else {
		archetype := desc.Archetype()
		row := desc.ArchetypeRow()

		changes.EachType(func(id internal.TypeID) {
			storage := s.registry.Storage(id)
			newColumnID := archetype.ComponentColumnID(id)
			stageColumnID := s.stager.ComponentColumnID(id)
			storage.CopyCell(newColumnID, row, stageColumnID, cmd.StageRow)
		})
	}
}

func (s *Scene) processComponentCommands(mask internal.TypeMask) (internal.TypeMask, internal.TypeMask) {
	changes := internal.EmptyTypeMask()

commandLoop:
	for s.commandBuffer.HasMoreData() {
		header := internal.ReadFromBuffer[internal.CommandHeader](s.commandBuffer)
		switch header.CommandType {

		case internal.CommandTypeAddComponent:
			cmd := internal.ReadFromBuffer[internal.AddComponentCommand](s.commandBuffer)
			if mask.HasType(cmd.TypeID) {
				panic(fmt.Errorf("cannot add component of type %v that the entity already has", cmd.TypeID))
			}
			mask.AddType(cmd.TypeID)
			changes.AddType(cmd.TypeID)

		case internal.CommandTypeRemoveComponent:
			cmd := internal.ReadFromBuffer[internal.RemoveComponentCommand](s.commandBuffer)
			if !mask.HasType(cmd.TypeID) {
				panic(fmt.Errorf("cannot remove component of type %v that the entity does not have", cmd.TypeID))
			}
			mask.RemoveType(cmd.TypeID)
			changes.RemoveType(cmd.TypeID)

		case internal.CommandTypeReplaceComponent:
			cmd := internal.ReadFromBuffer[internal.ReplaceComponentCommand](s.commandBuffer)
			if !mask.HasType(cmd.TypeID) {
				panic(fmt.Errorf("cannot replace component of type %v that the entity does not have", cmd.TypeID))
			}
			changes.AddType(cmd.TypeID)

		case internal.CommandTypeEndOfSequence:
			break commandLoop

		default:
			panic(fmt.Errorf("unexpected command type %v in EditEntity command buffer", header.CommandType))
		}
	}

	return mask, changes
}

func (s *Scene) processDeleteEntityCommand(cmd internal.DeleteEntityCommand) {
	id := fromInternalID(cmd.EntityID)

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic(fmt.Errorf("entity with ID %v does not exist", id))
	}

	oldMask := desc.Archetype().TypeMask()

	s.dispatchExitEvent(id, oldMask, internal.EmptyTypeMask())

	s.releaseArchetypeRow(desc.Destroy())
	s.releaseEntityIndex(id.index)
}

func (s *Scene) dispatchExitEvent(id ID, oldMask, newMask internal.TypeMask) {
	for subscription := range s.exitSubscriptions.CallbacksIter() {
		condition := subscription.condition
		if !condition.isSatisfiedBy(oldMask) {
			continue
		}
		if condition.isSatisfiedBy(newMask) {
			continue
		}
		subscription.callback(id)
	}
}

func (s *Scene) dispatchEnterEvent(id ID, oldMask opt.T[internal.TypeMask], newMask internal.TypeMask) {
	for subscription := range s.enterSubscriptions.CallbacksIter() {
		condition := subscription.condition
		if oldMask.Specified && condition.isSatisfiedBy(oldMask.Value) {
			continue
		}
		if !condition.isSatisfiedBy(newMask) {
			continue
		}
		subscription.callback(id)
	}
}
