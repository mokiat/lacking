package ecs

import "github.com/mokiat/lacking/util/observer"

func newScene() *Scene {
	scene := &Scene{
		deleteSubscriptions: observer.NewSubscriptionSet[DeleteCallback](),
	}
	resultCache := make([]*Result, 0, 3)
	for i := range resultCache {
		resultCache[i] = newResult(scene)
	}
	scene.resultCache = resultCache
	return scene
}

// Scene represents a collection of ECS entities.
type Scene struct {
	firstEntity  *Entity
	lastEntity   *Entity
	cachedEntity *Entity

	resultCache []*Result

	deleteSubscriptions *observer.SubscriptionSet[DeleteCallback]
}

// SubscribeDelete adds a callback to be executed before an entity is fully
// deleted.
func (s *Scene) SubscribeDelete(callback DeleteCallback) *DeleteSubscription {
	return s.deleteSubscriptions.Subscribe(callback)
}

// CreateEntity creates a new ECS entity in this
// scene.
func (s *Scene) CreateEntity() *Entity {
	var entity *Entity
	if s.cachedEntity != nil {
		entity = s.cachedEntity
		s.cachedEntity = s.cachedEntity.next
	} else {
		entity = &Entity{}
	}
	entity.scene = s
	entity.prev = nil
	entity.next = nil
	s.attachEntity(entity)
	return entity
}

// Find performs a search over the entities in
// this scene.
func (s *Scene) Find(query Query) *Result {
	var result *Result
	if count := len(s.resultCache); count > 0 {
		result = s.resultCache[count-1]
		s.resultCache = s.resultCache[:count-1]
	} else {
		result = newResult(s)
	}
	for entity := s.firstEntity; entity != nil; entity = entity.next {
		if entity.matches(query) {
			result.entities = append(result.entities, entity)
		}
	}
	return result
}

// Delete removes this scene and releases any
// allocated resources.
func (s *Scene) Delete() {
	s.firstEntity = nil
	s.lastEntity = nil
	s.cachedEntity = nil
}

func (s *Scene) notifyDelete(entity *Entity) {
	for callback := range s.deleteSubscriptions.CallbacksIter() {
		callback(entity)
	}
}

func (s *Scene) attachEntity(entity *Entity) {
	if s.firstEntity == nil {
		s.firstEntity = entity
	}
	if s.lastEntity != nil {
		s.lastEntity.next = entity
		entity.prev = s.lastEntity
	}
	entity.next = nil
	s.lastEntity = entity
}

func (s *Scene) detachEntity(entity *Entity) {
	if s.firstEntity == entity {
		s.firstEntity = entity.next
	}
	if s.lastEntity == entity {
		s.lastEntity = entity.prev
	}
	if entity.next != nil {
		entity.next.prev = entity.prev
	}
	if entity.prev != nil {
		entity.prev.next = entity.next
	}
	entity.prev = nil
	entity.next = nil
}

func (s *Scene) cacheEntity(entity *Entity) {
	entity.next = s.cachedEntity
	s.cachedEntity = entity
}

func (s *Scene) cacheResult(result *Result) {
	if len(s.resultCache) == cap(s.resultCache) {
		return
	}
	s.resultCache = append(s.resultCache, result)
}
