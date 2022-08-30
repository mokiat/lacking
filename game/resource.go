package game

import (
	"errors"
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/game/asset"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrStillLoading = errors.New("resource still loading")
)

func newResourceSet(parent *ResourceSet, engine *Engine) *ResourceSet {
	return &ResourceSet{
		parent:    parent,
		engine:    engine,
		registry:  engine.registry,
		ioWorker:  engine.ioWorker,
		gfxWorker: engine.gfxWorker,

		namedTwoDTextures: make(map[string]async.Promise[*TwoDTexture]),
		namedCubeTextures: make(map[string]async.Promise[*CubeTexture]),
		namedModels:       make(map[string]async.Promise[*ModelDefinition]),
	}
}

type ResourceSet struct {
	parent    *ResourceSet
	engine    *Engine
	registry  asset.Registry
	ioWorker  Worker
	gfxWorker Worker

	namedTwoDTextures map[string]async.Promise[*TwoDTexture]
	namedCubeTextures map[string]async.Promise[*CubeTexture]
	namedModels       map[string]async.Promise[*ModelDefinition]
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
		texture, err := s.allocateTwoDTexture(resource)
		if err != nil {
			result.Fail(fmt.Errorf("error loading twod texture %q: %w", id, err))
		} else {
			result.Deliver(texture)
		}
	}()
	s.namedTwoDTextures[id] = result
	return result
}

func (s *ResourceSet) OpenCubeTexture(id string) async.Promise[*CubeTexture] {
	if result, ok := s.findCubeTexture(id); ok {
		return result
	}

	resource := s.registry.ResourceByID(id)
	if resource == nil {
		return async.NewFailedPromise[*CubeTexture](fmt.Errorf("%w: %q", ErrNotFound, id))
	}

	result := async.NewPromise[*CubeTexture]()
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
		model, err := s.allocateModel(resource)
		if err != nil {
			result.Fail(fmt.Errorf("error loading model %q: %w", id, err))
		} else {
			result.Deliver(model)
		}
	}()
	s.namedModels[id] = result
	return result
}

// Delete schedules all resources managed by this ResourceSet for deletion.
// After this method returns, the resources are not guaranteed to have been
// released.
//
// Calling this method twice is not allowed. Allocating new resources after this
// method has been called is also not allowed.
func (s *ResourceSet) Delete() {
	go func() {
		for _, promise := range s.namedTwoDTextures {
			if texture, err := promise.Wait(); err == nil {
				s.releaseTwoDTexture(texture)
			}
		}
		s.namedTwoDTextures = nil
		for _, promise := range s.namedCubeTextures {
			if texture, err := promise.Wait(); err == nil {
				s.releaseCubeTexture(texture)
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

func (s *ResourceSet) findCubeTexture(id string) (async.Promise[*CubeTexture], bool) {
	if result, ok := s.namedCubeTextures[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findCubeTexture(id)
	}
	return async.Promise[*CubeTexture]{}, false
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
