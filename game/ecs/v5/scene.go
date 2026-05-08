package ecs

import (
	"iter"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/ecs/v5/internal"
)

// NewScene creates and initializes a new scene.
func NewScene(scope *Scope) *Scene {
	return &Scene{
		registry: scope.registry,

		freeEntityIndices: ds.EmptyStack[uint32](),
		entities:          nil,

		archetypePool: ds.EmptyStack[*internal.Archetype](),
		archetypes:    make(map[internal.TypeMask]*internal.Archetype),
	}
}

// Scene represents a collection of entities and their components. It provides
// methods for creating, deleting, and querying entities, as well as subscribing
// to entity events.
type Scene struct {
	registry *internal.Registry

	freeEntityIndices *ds.Stack[uint32]
	entities          []internal.Entity

	archetypePool *ds.Stack[*internal.Archetype]
	archetypes    map[internal.TypeMask]*internal.Archetype

	editOperation    EditOperation
	readOperation    ReadOperation
	inOperationBlock bool
}

// CreateEntity creates a new empty entity in the scene and returns its ID.
func (s *Scene) CreateEntity() ID {
	s.verifyOutsideOperation()

	index := s.allocateEntityIndex()
	desc := &s.entities[index]
	desc.Revive(index)
	desc.Assign(s.borrowArchetypeRow(internal.EmptyTypeMask()))

	return ID{
		index:    index,
		revision: desc.Revision(),
	}
}

// DeleteEntity deletes the entity from the scene. The entity is first stripped
// of all its components and then marked as invalid.
//
// Any references to components previously held by the entity must not be used
// after this call.
func (s *Scene) DeleteEntity(entityID ID) {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(entityID)
	if !ok {
		panic("entity does not exist")
	}

	s.releaseArchetypeRow(desc.Destroy())
	s.releaseEntityIndex(entityID.index)
}

// HasEntity returns whether the scene contains the specified entity.
func (s *Scene) HasEntity(entityID ID) bool {
	s.verifyOutsideOperation()

	_, ok := s.getEntityDescriptor(entityID)
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
func (s *Scene) ReadEntity(entity ID, fn func(*ReadOperation)) {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(entity)
	if !ok {
		panic("entity does not exist")
	}

	archetype := desc.Archetype()
	mask := archetype.TypeMask()
	row := desc.ArchetypeRow()

	columns, lookup := archetype.ComponentColumns()

	s.readOperation = ReadOperation{
		mask:             mask,
		row:              row,
		componentLookup:  lookup,
		componentColumns: columns,
	}
	s.inOperationBlock = true
	fn(&s.readOperation)
	s.inOperationBlock = false
}

// EditEntity allows editing the components of an entity.
//
// The provided callback will be called with an EntityChange that can be used
// to specify the changes to be applied to the entity. Trying to remove a
// component and adding it back in the same edit and vice versa is not
// supported and has undefined behavior.
func (s *Scene) EditEntity(id ID, fn EditOperationFunc) {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic("entity does not exist")
	}

	oldMask := desc.Archetype().TypeMask()
	oldArchetype := desc.Archetype()
	oldRow := desc.ArchetypeRow()

	s.editOperation = EditOperation{
		mask: oldMask,
	}
	s.inOperationBlock = true
	fn(&s.editOperation)
	s.inOperationBlock = false

	newMask := s.editOperation.mask
	if newMask == oldMask {
		return // no changes to apply
	}

	newArchetype, newRow := s.borrowArchetypeRow(newMask)

	newMask.EachType(func(id internal.TypeID) {
		newColumn := newArchetype.ComponentColumn(id)
		if oldMask.HasType(id) {
			oldColumn := oldArchetype.ComponentColumn(id)
			newColumn.CopyFromColumn(newRow, oldColumn, oldRow)
		} else {
			newColumn.CopyFromStorage(newRow)
		}
	})

	s.releaseArchetypeRow(oldArchetype, oldRow)

	desc.Assign(newArchetype, newRow)

	// 	// TODO: call subscribers.
	// 	// TODO: Abort this process if a subscriber made changes to the entity
}

// QueryEntities queries the scene for entities satisfying the specified
// condition and invokes the callback for each of them with a ReadOperation
// that can be used to read the components of the entity.
func (s *Scene) QueryEntities(condition Condition, yield func(ID, *ReadOperation) bool) {
	// TODO: If nested queries are supported, this method needs to be able to
	// fetch read operations from a pool.
	readOperation := &s.readOperation

	for mask, archetype := range s.archetypes {
		if !condition.isSatisfiedBy(mask) {
			continue
		}

		// TODO: Freeze archetype.

		columns, lookup := archetype.ComponentColumns()

		readOperation.mask = mask
		readOperation.componentLookup = lookup
		readOperation.componentColumns = columns

		for row := internal.Row(0); uint32(row) < archetype.Size(); row++ {
			intID := archetype.IDColumn().Value(row)
			if intID == internal.NilID {
				continue
			}

			readOperation.row = row

			s.inOperationBlock = true
			if !yield(fromInternalID(intID), &s.readOperation) {
				s.inOperationBlock = false
				return
			}
			s.inOperationBlock = false
		}

		// TODO: Unfreeze archetype.
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

// // SubscribeStartsSatisfying registers a callback that will be called whenever
// // an entity starts satisfying the specified condition. The callback will be
// // called with the ID of the entity that started satisfying the condition.
// // The returned subscription can be used to unsubscribe the callback.
// func (s *Scene) SubscribeStartsSatisfying(condition Condition, callback EntityCallback) *EntitySubscription {
// 	panic("not implemented")
// }

// // SubscribeStopsSatisfying registers a callback that will be called whenever
// // an entity stops satisfying the specified condition. The callback will be
// // called with the ID of the entity that stopped satisfying the condition.
// // The returned subscription can be used to unsubscribe the callback.
// func (s *Scene) SubscribeStopsSatisfying(condition Condition, callback EntityCallback) *EntitySubscription {
// 	panic("not implemented")
// }

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
