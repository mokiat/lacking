package ecs5

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/seq"
	"github.com/mokiat/lacking/util/mem"
	"github.com/mokiat/lacking/util/observer"
)

const defaultMaxEntityCount = 1024 * 1024

func newScene(maxEntityCount int) *Scene {
	freeEntityIndices := ds.NewStack[uint32](maxEntityCount)
	for i := range seq.Range(maxEntityCount-1, 0) {
		freeEntityIndices.Push(uint32(i))
	}

	return &Scene{
		deleteSubscriptions: observer.NewSubscriptionSet[DeleteCallback](),
		maxEntityCount:      maxEntityCount,
		entityMask:          newBitmask(),
		freeEntityIndices:   freeEntityIndices,
		handles:             make([]entityHandle, maxEntityCount),
		freeRevision:        uint32(1),
		results:             mem.NewSparseAllocator[Result](),
	}
}

// Scene represents a collection of ECS entities.
type Scene struct {
	deleteSubscriptions *observer.SubscriptionSet[DeleteCallback]
	maxEntityCount      int
	entityMask          *bitmask
	freeEntityIndices   *ds.Stack[uint32]
	handles             []entityHandle
	freeRevision        uint32
	freeComponentIndex  uint64
	results             *mem.SparseAllocator[Result]
}

// MaxEntityCount returns the maximum number of entities that can be managed
// by this scene at any given point in time (this includes entities marked
// for deletion that have not been purged yet).
func (s *Scene) MaxEntityCount() int {
	return s.maxEntityCount
}

// SubscribeDelete adds a callback to be executed before an entity is fully
// deleted.
func (s *Scene) SubscribeDelete(callback DeleteCallback) *DeleteSubscription {
	return s.deleteSubscriptions.Subscribe(callback)
}

// CreateEntity creates a new entity in this scene.
func (s *Scene) CreateEntity() Entity {
	s.freeRevision++
	index := s.freeEntityIndices.Pop()
	s.entityMask.Set(index, true)
	s.handles[index] = entityHandle{
		components:        componentMask(0),
		revision:          s.freeRevision,
		isPendingDeletion: false,
	}
	return Entity{
		scene:    s,
		index:    index,
		revision: s.freeRevision,
	}
}

// HasEntity returns whether the specified entity is still valid and
// part of this scene (i.e. it has not been marked for deletion and purged).
func (s *Scene) HasEntity(entity Entity) bool {
	handle := &s.handles[entity.index]
	return handle.revision == entity.revision
}

// DeleteEntity marks an entity for deletion.
func (s *Scene) DeleteEntity(entity Entity) {
	handle := &s.handles[entity.index]
	if handle.revision != entity.revision {
		return
	}
	handle.isPendingDeletion = true
}

// Query searches for entities that satisfy all specified conditions.
func (s *Scene) Query(conditions ...Condition) *Result {
	var queryCondition Condition
	for _, condition := range conditions {
		queryCondition.apply(condition)
	}
	result := s.results.Allocate()
	result.allocator = s.results
	result.entities = result.entities[:0]
	for entityIndex := range s.entityMask.ActiveIter() {
		if handle := &s.handles[entityIndex]; queryCondition.isSatisfied(handle) {
			result.entities = append(result.entities, Entity{
				scene:    s,
				index:    entityIndex,
				revision: handle.revision,
			})
		}
	}
	return result
}

// Purge removes any entities that have been marked for deletion.
//
// All delete subscriptions will be notified at this point in time.
func (s *Scene) Purge() {
	for entityIndex := range s.entityMask.ActiveIter() {
		if handle := &s.handles[entityIndex]; handle.isPendingDeletion {
			s.notifyDelete(Entity{
				scene:    s,
				index:    entityIndex,
				revision: handle.revision,
			})
			s.entityMask.Set(entityIndex, false)
			s.freeEntityIndices.Push(entityIndex)
		}
	}
}

// Delete removes this scene and releases any associated resources.
func (s *Scene) Delete() {}

func (s *Scene) newComponentType() componentMask {
	if s.freeComponentIndex >= MaxComponentCount {
		panic("max number of components reached")
	}
	result := componentMask(1 << s.freeComponentIndex)
	s.freeComponentIndex++
	return result
}

func (s *Scene) notifyDelete(entity Entity) {
	for callback := range s.deleteSubscriptions.CallbacksIter() {
		callback(entity)
	}
}
