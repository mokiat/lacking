package ecs

import (
	"fmt"
	"iter"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/ecs/v6/internal"
	"github.com/mokiat/lacking/util/observer"
)

// NewScene creates and initializes a new scene.
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

// Scene represents a collection of entities and their components. It provides
// methods for creating, deleting, and querying entities, as well as subscribing
// to entity events.
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
	queryNesting      uint32
	inOperationBlock  bool
	inQueueProcessing bool
}

// Destroy destroys the scene and releases all associated resources. After
// calling this method, the scene must not be used anymore.
func (s *Scene) Destroy() {
	s.enterSubscriptions.Clear()
	s.exitSubscriptions.Clear()
	for _, archetype := range s.archetypes {
		archetype.Destroy()
	}
	s.stager.Destroy()
}

// SubscribeEnter registers a callback that will be called whenever
// an entity starts satisfying the specified condition. The callback will be
// called with the ID of the entity that started satisfying the condition.
// The returned subscription can be used to unsubscribe the callback.
func (s *Scene) SubscribeEnter(condition Condition, callback EntityCallback) *EntitySubscription {
	return s.enterSubscriptions.Subscribe(ConditionalCallback{
		condition: condition,
		callback:  callback,
	})
}

// SubscribeExit registers a callback that will be called whenever
// an entity stops satisfying the specified condition. The callback will be
// called with the ID of the entity that stopped satisfying the condition.
// The returned subscription can be used to unsubscribe the callback.
func (s *Scene) SubscribeExit(condition Condition, callback EntityCallback) *EntitySubscription {
	return s.exitSubscriptions.Subscribe(ConditionalCallback{
		condition: condition,
		callback:  callback,
	})
}

// CreateEntity creates a new empty entity in the scene and returns its ID.
func (s *Scene) CreateEntity() ID {
	s.verifyOutsideOperation()
	s.verifyOutsideQuery()

	index := s.allocateEntityIndex()
	desc := &s.entities[index]
	desc.Revive(index)

	internal.WriteToBuffer(s.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeCreateEntity,
	})
	internal.WriteToBuffer(s.commandBuffer, internal.CreateEntityCommand{
		EntityID: desc.ID(),
	})

	if !s.inQueueProcessing {
		s.processQueue()
	}

	return fromInternalID(desc.ID())
}

// DeleteEntity deletes the entity from the scene. The entity is first stripped
// of all its components and then marked as invalid.
//
// Any references to components previously held by the entity must not be used
// after this call.
func (s *Scene) DeleteEntity(id ID) {
	s.verifyOutsideOperation()
	s.verifyOutsideQuery()

	internal.WriteToBuffer(s.commandBuffer, internal.CommandHeader{
		CommandType: internal.CommandTypeDeleteEntity,
	})
	internal.WriteToBuffer(s.commandBuffer, internal.DeleteEntityCommand{
		EntityID: internal.NewID(id.index, id.revision),
	})

	if !s.inQueueProcessing {
		s.processQueue()
	}
}

// HasEntity returns whether the scene contains the specified entity.
func (s *Scene) HasEntity(id ID) bool {
	s.verifyOutsideOperation()

	_, ok := s.getEntityDescriptor(id)
	return ok
}

// CheckEntity returns whether the specified entity satisfies the specified
// condition.
//
// This method does allow for invalid or deleted entity IDs to be passed in,
// and will simply return false for them.
func (s *Scene) CheckEntity(id ID, condition Condition) bool {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		return false
	}
	archetype := desc.Archetype()

	return condition.isSatisfiedBy(archetype.TypeMask())
}

// ReadEntity allows reading the components of an entity.
//
// The provided callback will be called with a ReadOperation that can be used
// to specify the components to be read from the entity.
func (s *Scene) ReadEntity(id ID, fn func(*ReadOperation)) {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic("entity does not exist")
	}

	archetype := desc.Archetype()
	mask := archetype.TypeMask()
	row := desc.ArchetypeRow()
	columns, lookup := archetype.ComponentColumns()

	readOperation := s.allocateReadOperation()
	defer s.releaseReadOperation(readOperation)
	readOperation.mask = mask
	readOperation.componentLookup = lookup
	readOperation.componentColumns = columns
	readOperation.row = row

	s.inOperationBlock = true
	fn(readOperation)
	s.inOperationBlock = false
}

// EditEntity allows editing the components of an entity.
//
// The provided callback will be called with an EditOperation that can be used
// to specify the changes to be applied to the entity.
//
// Trying to add a component that the entity already has or remove a component
// that the entity does not have will result in a panic.
//
// If multiple changes are made to the same component type, only the last one
// will be applied, and the intermediate ones will be ignored.
func (s *Scene) EditEntity(id ID, fn EditOperationFunc) {
	s.verifyOutsideOperation()
	s.verifyOutsideQuery()

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

	if !s.inQueueProcessing {
		s.processQueue()
	}
}

// QueryEntities queries the scene for entities satisfying the specified
// condition and invokes the callback for each of them with a ReadOperation
// that can be used to read the components of the entity.
func (s *Scene) QueryEntities(condition Condition, yield func(ID, *ReadOperation) bool) {
	s.verifyOutsideOperation()

	s.queryNesting++
	defer func() {
		s.queryNesting--
	}()

	readOperation := s.allocateReadOperation()
	defer s.releaseReadOperation(readOperation)

	for mask, archetype := range s.archetypes {
		if !condition.isSatisfiedBy(mask) {
			continue
		}

		columns, lookup := archetype.ComponentColumns()

		readOperation.mask = mask
		readOperation.componentLookup = lookup
		readOperation.componentColumns = columns

		for row := internal.Row(0); uint32(row) < archetype.Size(); row++ {
			readOperation.row = row
			intID := archetype.IDColumn().Value(row)
			if !yield(fromInternalID(intID), readOperation) {
				return
			}
		}
	}
}

// QueryEntities queries the scene for entities satisfying the specified
// condition and invokes the callback for each of them with a ReadOperation
// that can be used to read the components of the entity.
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

func (s *Scene) verifyOutsideQuery() {
	if s.queryNesting > 0 {
		panic("cannot call this method during queries")
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
	if s.inQueueProcessing {
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

	oldMask := opt.Unspecified[internal.TypeMask]() // from limbo
	newMask := internal.EmptyTypeMask()             // empty archetype

	desc.Assign(s.borrowArchetypeRow(newMask))

	s.dispatchEnterEvent(id, oldMask, newMask)
}

func (s *Scene) processEditEntityCommand(cmd internal.EditEntityCommand) {
	id := fromInternalID(cmd.EntityID)

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic(fmt.Errorf("entity with ID %v does not exist", id))
	}

	oldMask := desc.Archetype().TypeMask()
	newMask := oldMask
	changes := internal.EmptyTypeMask()

commandLoop:
	for s.commandBuffer.HasMoreData() {
		header := internal.ReadFromBuffer[internal.CommandHeader](s.commandBuffer)
		switch header.CommandType {

		case internal.CommandTypeAddComponent:
			cmd := internal.ReadFromBuffer[internal.AddComponentCommand](s.commandBuffer)
			if newMask.HasType(cmd.TypeID) {
				panic(fmt.Errorf("cannot add component of type %v that the entity already has", cmd.TypeID))
			}
			newMask.AddType(cmd.TypeID)
			changes.AddType(cmd.TypeID)

		case internal.CommandTypeRemoveComponent:
			cmd := internal.ReadFromBuffer[internal.RemoveComponentCommand](s.commandBuffer)
			if !newMask.HasType(cmd.TypeID) {
				panic(fmt.Errorf("cannot remove component of type %v that the entity does not have", cmd.TypeID))
			}
			newMask.RemoveType(cmd.TypeID)
			changes.RemoveType(cmd.TypeID)

		case internal.CommandTypeReplaceComponent:
			cmd := internal.ReadFromBuffer[internal.ReplaceComponentCommand](s.commandBuffer)
			if !newMask.HasType(cmd.TypeID) {
				panic(fmt.Errorf("cannot replace component of type %v that the entity does not have", cmd.TypeID))
			}
			changes.AddType(cmd.TypeID)

		case internal.CommandTypeEndOfSequence:
			break commandLoop

		default:
			panic(fmt.Errorf("unexpected command type %v in EditEntity command buffer", header.CommandType))
		}
	}

	if oldMask != newMask {
		s.dispatchExitEvent(id, oldMask, newMask)

		oldArchetype := desc.Archetype()
		oldRow := desc.ArchetypeRow()

		newArchetype, newRow := s.borrowArchetypeRow(newMask)
		newMask.EachType(func(id internal.TypeID) {
			newColumn := newArchetype.ComponentColumn(id)
			if changes.HasType(id) {
				stageColumn := s.stager.ComponentColumn(id)
				newColumn.CopyFromColumn(newRow, stageColumn, cmd.StageRow)
			} else {
				oldColumn := oldArchetype.ComponentColumn(id)
				newColumn.CopyFromColumn(newRow, oldColumn, oldRow)
			}
		})
		s.releaseArchetypeRow(oldArchetype, oldRow)

		desc.Assign(newArchetype, newRow)

		s.dispatchEnterEvent(id, opt.V(oldMask), newMask)
	} else {
		changes.EachType(func(id internal.TypeID) {
			newColumn := desc.Archetype().ComponentColumn(id)
			stageColumn := s.stager.ComponentColumn(id)
			newColumn.CopyFromColumn(desc.ArchetypeRow(), stageColumn, cmd.StageRow)
		})
	}
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
