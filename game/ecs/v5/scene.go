package ecs

import (
	"github.com/mokiat/gog/ds"
)

// NewScene creates and initializes a new scene.
func NewScene() *Scene {
	return &Scene{
		freeEntityIndices: ds.EmptyStack[int32](),
		entities:          nil,

		archetypePool: ds.EmptyStack[*componentArchetype](),
		archetypes:    make(map[componentMask]*componentArchetype),
	}
}

// Scene represents a collection of entities and their components. It provides
// methods for creating, deleting, and querying entities, as well as subscribing
// to entity events.
type Scene struct {
	freeEntityIndices *ds.Stack[int32]
	entities          []entityDescriptor

	archetypePool *ds.Stack[*componentArchetype]
	archetypes    map[componentMask]*componentArchetype

	inOperationBlock bool
}

// CreateEntity creates a new empty entity in the scene and returns its ID.
func (s *Scene) CreateEntity() EntityID {
	s.verifyOutsideOperation()

	index := s.allocateEntityIndex()
	desc := &s.entities[index]
	desc.revision++
	desc.archetype, desc.archetypeOffset = s.borrowArchetypeSlot(emptyComponentMask())

	return EntityID{
		index:    index,
		revision: desc.revision,
	}
}

// DeleteEntity deletes the entity from the scene. The entity is first stripped
// of all its components and then marked as invalid.
//
// Any references to components previously held by the entity must not be used
// after this call.
func (s *Scene) DeleteEntity(entityID EntityID) {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(entityID)
	if !ok {
		panic("entity does not exist")
	}

	s.releaseArchetypeSlot(desc.archetype, desc.archetypeOffset)
	desc.revision++
	desc.archetype = nil
	desc.archetypeOffset = 0

	s.releaseEntityIndex(entityID.index)
}

// HasEntity returns whether the scene contains the specified entity.
func (s *Scene) HasEntity(entityID EntityID) bool {
	s.verifyOutsideOperation()

	_, ok := s.getEntityDescriptor(entityID)
	return ok
}

// CheckEntity returns whether the specified entity satisfies the specified
// condition.
//
// This method does allow for invalid or deleted entity IDs to be passed in,
// and will simply return false for them.
func (s *Scene) CheckEntity(id EntityID, condition Condition) bool {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		return false
	}
	archetype := desc.archetype

	return condition.isSatisfiedBy(archetype.mask)
}

// ReadEntity allows reading the components of an entity.
//
// The provided callback will be called with a ReadOperation that can be used
// to specify the components to be read from the entity.
func (s *Scene) ReadEntity(entity EntityID, fn func(*ReadOperation)) {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(entity)
	if !ok {
		panic("entity does not exist")
	}

	op := ReadOperation{
		scene:           s,
		archetype:       desc.archetype,
		archetypeOffset: desc.archetypeOffset,
	}
	s.inOperationBlock = true
	fn(&op)
	s.inOperationBlock = false
}

// EditEntity allows editing the components of an entity.
//
// The provided callback will be called with an EntityChange that can be used
// to specify the changes to be applied to the entity. Trying to remove a
// component and adding it back in the same edit and vice versa is not
// supported and has undefined behavior.
func (s *Scene) EditEntity(id EntityID, fn EditOperationFunc) {
	s.verifyOutsideOperation()

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic("entity does not exist")
	}

	oldMask := desc.archetype.mask

	change := EditOperation{
		scene: s,
		mask:  oldMask,
	}
	s.inOperationBlock = true
	fn(&change)
	s.inOperationBlock = false

	newMask := change.mask

	if newMask == oldMask {
		return // no changes to apply
	}

	oldArchetype := desc.archetype
	oldOffset := desc.archetypeOffset

	desc.archetype, desc.archetypeOffset = s.borrowArchetypeSlot(newMask)
	// 	// TODO: relocate entity
	s.releaseArchetypeSlot(oldArchetype, oldOffset)

	// 	// TODO: call subscribers.
	// 	// TODO: Abort this process if a subscriber made changes to the entity
}

// // QueryEntities queries the scene for entities satisfying the specified
// // condition and invokes the callback for each of them with a ReadOperation
// // that can be used to read the components of the entity.
// func (s *Scene) QueryEntities(condition Condition, cb func(EntityID, *ReadOperation)) {
// 	panic("not implemented")
// }

// // QueryEntities queries the scene for entities satisfying the specified
// // condition and invokes the callback for each of them with a ReadOperation
// // that can be used to read the components of the entity.
// func (s *Scene) QueryEntitiesIter(condition Condition) iter.Seq2[EntityID, *ReadOperation] {
// 	panic("not implemented")
// }

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

func (s *Scene) allocateEntityIndex() int32 {
	var index int32
	if s.freeEntityIndices.IsEmpty() {
		index = int32(len(s.entities))
		s.entities = append(s.entities, entityDescriptor{})
	} else {
		index = s.freeEntityIndices.Pop()
	}
	return index
}

func (s *Scene) releaseEntityIndex(index int32) {
	s.freeEntityIndices.Push(index)
}

func (s *Scene) getEntityDescriptor(id EntityID) (*entityDescriptor, bool) {
	if id == NilEntityID {
		return nil, false
	}
	desc := &s.entities[id.index]
	if desc.revision != id.revision {
		return nil, false
	}
	return desc, true
}

func (s *Scene) borrowArchetypeSlot(mask componentMask) (*componentArchetype, uint32) {
	archetype, ok := s.archetypes[mask]
	if !ok {
		archetype = s.allocateArchetype()
		s.archetypes[mask] = archetype

		archetype.mask = mask
		// for tIndex := range mask.typeIndicesIter() {
		// 	storage := s.storages[tIndex]
		// 	archetype.components[tIndex] = storage.createChain()
		// }
	}

	offset := archetype.allocateOffset()
	return archetype, offset
}

func (s *Scene) releaseArchetypeSlot(archetype *componentArchetype, offset uint32) {
	// panic("not implemented")
}

func (s *Scene) allocateArchetype() *componentArchetype {
	if !s.archetypePool.IsEmpty() {
		return s.archetypePool.Pop()
	}
	result := new(componentArchetype)
	result.reset()
	return result
}

func (s *Scene) releaseArchetype(archetype *componentArchetype) {
	archetype.reset()
	s.archetypePool.Push(archetype)
}
