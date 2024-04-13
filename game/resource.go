package game

import (
	"errors"
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	newasset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/async"
)

// TOOD: Only the game resource sets should manage resource set hierarchies.
// The graphics, physics, audio packages should expose resource creation
// through their engine's and their scenes should only serve as a
// grouping/switch mechanism.

var ErrNotFound = errors.New("resource not found")

func newResourceSet(parent *ResourceSet, engine *Engine) *ResourceSet {
	return &ResourceSet{
		parent:      parent,
		renderAPI:   engine.Graphics().API(),
		engine:      engine,
		registry:    engine.registry,
		newRegistry: engine.newRegistry,
		ioWorker:    engine.ioWorker,
		gfxWorker:   engine.gfxWorker,

		namedModels:  make(map[string]async.Promise[*ModelDefinition]),
		namedModels2: make(map[string]async.Promise[SceneDefinition2]),
	}
}

type ResourceSet struct {
	parent      *ResourceSet
	renderAPI   render.API
	engine      *Engine
	registry    asset.Registry
	newRegistry *newasset.Registry
	ioWorker    Worker
	gfxWorker   Worker

	namedModels  map[string]async.Promise[*ModelDefinition]
	namedModels2 map[string]async.Promise[SceneDefinition2]
}

func (s *ResourceSet) CreateResourceSet() *ResourceSet {
	return newResourceSet(s, s.engine)
}

func (s *ResourceSet) OpenModel(id string) async.Promise[*ModelDefinition] {
	if result, ok := s.findModel(id); ok {
		return result
	}

	resource := s.registry.ResourceByID(id)
	if resource == nil {
		return async.NewFailedPromise[*ModelDefinition](fmt.Errorf("%w: %q", ErrNotFound, id))
	}

	result := async.NewPromise[*ModelDefinition]()
	go func() {
		model, err := s.loadModel(resource)
		if err != nil {
			result.Fail(fmt.Errorf("error loading model %q: %w", id, err))
		} else {
			result.Deliver(model)
		}
	}()
	s.namedModels[id] = result
	return result
}

func (s *ResourceSet) OpenModelByName(name string) async.Promise[*ModelDefinition] {
	resource := s.registry.ResourceByName(name)
	if resource == nil {
		return async.NewFailedPromise[*ModelDefinition](fmt.Errorf("%w: %q", ErrNotFound, name))
	}
	return s.OpenModel(resource.ID())
}

func (s *ResourceSet) OpenFragmentWithID(id string) async.Promise[SceneDefinition2] {
	if result, ok := s.findModel2(id); ok {
		return result
	}

	resource := s.newRegistry.ResourceByID(id)
	if resource == nil {
		return async.NewFailedPromise[SceneDefinition2](fmt.Errorf("%w: %q", ErrNotFound, id))
	}

	result := async.NewPromise[SceneDefinition2]()
	go func() {
		fragment, err := s.loadModel2(resource)
		if err != nil {
			result.Fail(fmt.Errorf("error loading model %q: %w", id, err))
		} else {
			result.Deliver(fragment)
		}
	}()
	return result
}

func (s *ResourceSet) OpenFragmentWithName(name string) async.Promise[SceneDefinition2] {
	resource := s.newRegistry.ResourceByName(name)
	if resource == nil {
		return async.NewFailedPromise[SceneDefinition2](fmt.Errorf("%w: %q", ErrNotFound, name))
	}
	return s.OpenFragmentWithID(resource.ID())
}

// Delete schedules all resources managed by this ResourceSet for deletion.
// After this method returns, the resources are not guaranteed to have been
// released.
//
// Calling this method twice is not allowed. Allocating new resources after this
// method has been called is also not allowed.
func (s *ResourceSet) Delete() {
	// FIXME: All of the release calls need to occur on the GPU thread.
	// Also, rework this method to return a promise.
	go func() {
		for _, promise := range s.namedModels {
			if model, err := promise.Wait(); err == nil {
				s.releaseModel(model)
			}
		}
		s.namedModels = nil

		// TODO: Release named models2
	}()
}

func (s *ResourceSet) findModel(id string) (async.Promise[*ModelDefinition], bool) {
	if result, ok := s.namedModels[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findModel(id)
	}
	return async.Promise[*ModelDefinition]{}, false
}

func (s *ResourceSet) findModel2(id string) (async.Promise[SceneDefinition2], bool) {
	if result, ok := s.namedModels2[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findModel2(id)
	}
	return async.Promise[SceneDefinition2]{}, false
}
