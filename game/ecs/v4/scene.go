package ecs

import (
	"iter"
	"reflect"

	"github.com/mokiat/gog/ds"
)

// NewScene creates and initializes a new scene.
func NewScene() *Scene {
	return &Scene{
		storageMapping: make(map[reflect.Type]typeIndex),
		storages:       nil,

		freeEntityIndices: ds.EmptyStack[int32](),
		entities:          nil,

		archetypePool: nil,
		archetypes:    make(map[componentMask]*componentArchetype),
	}
}

// Scene represents a collection of entities and their components. It provides
// methods for creating, deleting, and querying entities, as well as subscribing
// to entity events.
type Scene struct {
	storageMapping map[reflect.Type]typeIndex
	storages       []ComponentStorage

	freeEntityIndices *ds.Stack[int32]
	entities          []entityDescriptor

	archetypePool *ds.Pool[componentArchetype]
	archetypes    map[componentMask]*componentArchetype

	isEditing bool
}

// RegisterStorage registers a component storage, enabling the scene to manage
// components of the corresponding type.
func (s *Scene) RegisterStorage(storage ComponentStorage) {
	s.verifyNotEditing()

	reflectType := storage.reflectType()
	if _, ok := s.storageMapping[reflectType]; ok {
		panic("storage for this component type already registered")
	}
	s.storageMapping[reflectType] = typeIndex(len(s.storages))
	s.storages = append(s.storages, storage)
}

// CreateEntity creates a new empty entity in the scene and returns its ID.
func (s *Scene) CreateEntity() EntityID {
	s.verifyNotEditing()

	index := s.allocateEntityIndex()
	desc := &s.entities[index]
	desc.revision++
	desc.archetype, desc.archetypeOffset = s.enterArchetype(emptyComponentMask())

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
func (s *Scene) DeleteEntity(scene *Scene, entityID EntityID) {
	s.verifyNotEditing()

	desc, ok := s.getEntityDescriptor(entityID)
	if !ok {
		panic("entity does not exist")
	}

	s.leaveArchetype(desc.archetype, desc.archetypeOffset)
	desc.revision++
	desc.archetype = nil
	desc.archetypeOffset = 0

	s.releaseEntityIndex(entityID.index)
}

// HasEntity returns whether the scene contains the specified entity.
func (s *Scene) HasEntity(entityID EntityID) bool {
	s.verifyNotEditing()

	_, ok := s.getEntityDescriptor(entityID)
	return ok
}

// CheckEntity returns whether the specified entity satisfies the specified
// condition.
//
// This method does allow for invalid or deleted entity IDs to be passed in,
// and will simply return false for them.
func (s *Scene) CheckEntity(id EntityID, condition Condition) bool {
	s.verifyNotEditing()

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
	s.verifyNotEditing()

	desc, ok := s.getEntityDescriptor(entity)
	if !ok {
		panic("entity does not exist")
	}

	op := ReadOperation{
		scene:           s,
		archetype:       desc.archetype,
		archetypeOffset: desc.archetypeOffset,
	}
	s.isEditing = true
	fn(&op)
	s.isEditing = false
}

// EditEntity allows editing the components of an entity.
//
// The provided callback will be called with an EntityChange that can be used
// to specify the changes to be applied to the entity. Trying to remove a
// component and adding it back in the same edit and vice versa is not
// supported and has undefined behavior.
func (s *Scene) EditEntity(id EntityID, fn func(*EditOperation)) {
	s.verifyNotEditing()

	desc, ok := s.getEntityDescriptor(id)
	if !ok {
		panic("entity does not exist")
	}

	archetype := desc.archetype
	oldMask := archetype.mask

	change := EditOperation{
		scene: s,
		mask:  oldMask,
	}
	s.isEditing = true // TODO: Either rename to something else or split into different "is<Something>" flags
	fn(&change)
	s.isEditing = false

	newMask := change.mask
	if newMask != oldMask {
		// TODO: relocate entity
		// TODO: call subscribers.
		// TODO: Abort this process if a subscriber made changes to the entity
	}
}

// QueryEntities queries the scene for entities satisfying the specified
// condition and invokes the callback for each of them with a ReadOperation
// that can be used to read the components of the entity.
func (s *Scene) QueryEntities(condition Condition, cb func(EntityID, *ReadOperation)) {
	panic("not implemented")
}

// QueryEntities queries the scene for entities satisfying the specified
// condition and invokes the callback for each of them with a ReadOperation
// that can be used to read the components of the entity.
func (s *Scene) QueryEntitiesIter(condition Condition) iter.Seq2[EntityID, *ReadOperation] {
	panic("not implemented")
}

// SubscribeStartsSatisfying registers a callback that will be called whenever
// an entity starts satisfying the specified condition. The callback will be
// called with the ID of the entity that started satisfying the condition.
// The returned subscription can be used to unsubscribe the callback.
func (s *Scene) SubscribeStartsSatisfying(condition Condition, callback EntityCallback) *EntitySubscription {
	panic("not implemented")
}

// SubscribeStopsSatisfying registers a callback that will be called whenever
// an entity stops satisfying the specified condition. The callback will be
// called with the ID of the entity that stopped satisfying the condition.
// The returned subscription can be used to unsubscribe the callback.
func (s *Scene) SubscribeStopsSatisfying(condition Condition, callback EntityCallback) *EntitySubscription {
	panic("not implemented")
}

func (s *Scene) verifyNotEditing() {
	if s.isEditing {
		panic("cannot call this method during an edit")
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

func (s *Scene) enterArchetype(mask componentMask) (*componentArchetype, uint32) {
	archetype, ok := s.archetypes[mask]
	if !ok {
		archetype = s.archetypePool.Fetch()
		s.archetypes[mask] = archetype
		if archetype.components == nil {
			archetype.components = make(map[typeIndex]componentChain)
		}
	}

	archetype.mask = mask
	for tIndex := range mask.typeIndicesIter() {
		storage := s.storages[tIndex]
		archetype.components[tIndex] = storage.createChain()
	}

	offset := archetype.allocateOffset()
	return archetype, offset
}

func (s *Scene) leaveArchetype(archetype *componentArchetype, offset uint32) {
	panic("not implemented")
}
