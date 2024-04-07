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

var (
	ErrNotFound     = errors.New("resource not found")
	ErrStillLoading = errors.New("resource still loading")
)

func newResourceSet(parent *ResourceSet, engine *Engine) *ResourceSet {
	return &ResourceSet{
		parent:      parent,
		engine:      engine,
		registry:    engine.registry,
		newRegistry: engine.newRegistry,
		ioWorker:    engine.ioWorker,
		gfxWorker:   engine.gfxWorker,

		namedTwoDTextures: make(map[string]async.Promise[*TwoDTexture]),
		namedCubeTextures: make(map[string]async.Promise[render.Texture]),
		namedModels:       make(map[string]async.Promise[*ModelDefinition]),
		namedScenes:       make(map[string]async.Promise[*SceneDefinition]),
	}
}

type ResourceSet struct {
	parent      *ResourceSet
	engine      *Engine
	registry    asset.Registry
	newRegistry *newasset.Registry
	ioWorker    Worker
	gfxWorker   Worker

	namedTwoDTextures map[string]async.Promise[*TwoDTexture]
	namedCubeTextures map[string]async.Promise[render.Texture]
	namedModels       map[string]async.Promise[*ModelDefinition]
	namedScenes       map[string]async.Promise[*SceneDefinition]
}

func (s *ResourceSet) CreateResourceSet() *ResourceSet {
	return newResourceSet(s, s.engine)
}

func (s *ResourceSet) OpenTwoDTexture(id string) async.Promise[*TwoDTexture] {
	if result, ok := s.findTwoDTexture(id); ok {
		return result
	}

	resource := s.registry.ResourceByID(id)
	if resource == nil {
		return async.NewFailedPromise[*TwoDTexture](fmt.Errorf("%w: %q", ErrNotFound, id))
	}

	result := async.NewPromise[*TwoDTexture]()
	go func() {
		texture, err := s.loadTwoDTexture(resource)
		if err != nil {
			result.Fail(fmt.Errorf("error loading twod texture %q: %w", id, err))
		} else {
			result.Deliver(texture)
		}
	}()
	s.namedTwoDTextures[id] = result
	return result
}

func (s *ResourceSet) OpenCubeTexture(id string) async.Promise[render.Texture] {
	if result, ok := s.findCubeTexture(id); ok {
		return result
	}

	resource := s.registry.ResourceByID(id)
	if resource == nil {
		return async.NewFailedPromise[render.Texture](fmt.Errorf("%w: %q", ErrNotFound, id))
	}

	result := async.NewPromise[render.Texture]()
	go func() {
		texture, err := s.allocateCubeTexture(resource)
		if err != nil {
			result.Fail(fmt.Errorf("error loading cube texture %q: %w", id, err))
		} else {
			result.Deliver(texture)
		}
	}()
	s.namedCubeTextures[id] = result
	return result
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

func (s *ResourceSet) OpenScene(id string) async.Promise[*SceneDefinition] {
	if result, ok := s.findScene(id); ok {
		return result
	}

	resource := s.registry.ResourceByID(id)
	if resource == nil {
		return async.NewFailedPromise[*SceneDefinition](fmt.Errorf("%w: %q", ErrNotFound, id))
	}

	result := async.NewPromise[*SceneDefinition]()
	go func() {
		model, err := s.allocateScene(resource)
		if err != nil {
			result.Fail(fmt.Errorf("error loading level %q: %w", id, err))
		} else {
			result.Deliver(model)
		}
	}()
	s.namedScenes[id] = result
	return result
}

func (s *ResourceSet) OpenSceneByName(name string) async.Promise[*SceneDefinition] {
	resource := s.registry.ResourceByName(name)
	if resource == nil {
		return async.NewFailedPromise[*SceneDefinition](fmt.Errorf("%w: %q", ErrNotFound, name))
	}
	return s.OpenScene(resource.ID())
}

func (s *ResourceSet) OpenFragmentWithID(id string) async.Promise[SceneDefinition2] {
	// TODO: Use caching

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
		for _, promise := range s.namedScenes {
			if scene, err := promise.Wait(); err == nil {
				s.releaseScene(scene)
			}
		}
		s.namedScenes = nil
		for _, promise := range s.namedTwoDTextures {
			if texture, err := promise.Wait(); err == nil {
				s.releaseTwoDTexture(texture)
			}
		}
		s.namedTwoDTextures = nil
		for _, promise := range s.namedCubeTextures {
			if texture, err := promise.Wait(); err == nil {
				texture.Release()
			}
		}
		s.namedCubeTextures = nil
		for _, promise := range s.namedModels {
			if model, err := promise.Wait(); err == nil {
				s.releaseModel(model)
			}
		}
		s.namedModels = nil
	}()
}

func (s *ResourceSet) findTwoDTexture(id string) (async.Promise[*TwoDTexture], bool) {
	if result, ok := s.namedTwoDTextures[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findTwoDTexture(id)
	}
	return async.Promise[*TwoDTexture]{}, false
}

func (s *ResourceSet) findCubeTexture(id string) (async.Promise[render.Texture], bool) {
	if result, ok := s.namedCubeTextures[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findCubeTexture(id)
	}
	return async.Promise[render.Texture]{}, false
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

func (s *ResourceSet) findScene(id string) (async.Promise[*SceneDefinition], bool) {
	if result, ok := s.namedScenes[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findScene(id)
	}
	return async.Promise[*SceneDefinition]{}, false
}
