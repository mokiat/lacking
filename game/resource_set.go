package game

import (
	"fmt"
	"sync"

	"github.com/mokiat/lacking/util/async"
)

func newResourceSet(engine *Engine, registry *resourceRegistry) *ResourceSet {
	return &ResourceSet{
		engine:   engine,
		registry: registry,

		trackedResources: make(map[string]int),
	}
}

// ResourceSet is a collection of resources that are managed together.
//
// This makes it possible to release multiple resources at once, without
// having to track them individually.
type ResourceSet struct {
	engine   *Engine
	registry *resourceRegistry

	trackedResourcesMU sync.Mutex
	trackedResources   map[string]int
}

// Engine returns the engine associated with this ResourceSet.
func (s *ResourceSet) Engine() *Engine {
	return s.engine
}

// FetchResource requests a resource to be read from the storage and loaded.
//
// Once loaded, the resource will be stored in the target object and will
// be tracked by this ResourceSet. Once the ResourceSet is deleted, all
// tracked resources will be scheduled for deletion.
//
// This method can be called from any thread.
func (s *ResourceSet) FetchResource(path string, target any) async.Operation {
	s.trackedResourcesMU.Lock()
	defer s.trackedResourcesMU.Unlock()

	if s.trackedResources == nil {
		return async.NewFailedOperation(fmt.Errorf("resource set has been deleted"))
	}

	s.trackedResources[path]++
	return s.registry.LoadResource(s, path, target)
}

// Delete schedules all resources managed by this ResourceSet for deletion.
//
// This method can be called from any thread.
func (s *ResourceSet) Delete() {
	s.trackedResourcesMU.Lock()
	defer s.trackedResourcesMU.Unlock()

	for path, count := range s.trackedResources {
		s.registry.UnloadResource(s, path, count)
	}
	s.trackedResources = nil
}
