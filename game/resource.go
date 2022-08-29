package game

import (
	"errors"
	"fmt"

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

		namedTwoDTextures: make(map[string]Placeholder[*TwoDTexture]),
		namedCubeTextures: make(map[string]Placeholder[*CubeTexture]),
	}
}

type ResourceSet struct {
	parent    *ResourceSet
	engine    *Engine
	registry  asset.Registry
	ioWorker  Worker
	gfxWorker Worker

	namedTwoDTextures map[string]Placeholder[*TwoDTexture]
	namedCubeTextures map[string]Placeholder[*CubeTexture]
}

func (s *ResourceSet) CreateResourceSet() *ResourceSet {
	return newResourceSet(s, s.engine)
}

func (s *ResourceSet) OpenTwoDTexture(id string) Placeholder[*TwoDTexture] {
	if result, ok := s.findTwoDTexture(id); ok {
		return result
	}

	resource := s.registry.ResourceByID(id)
	if resource == nil {
		return failedPlaceholder[*TwoDTexture](fmt.Errorf("%w: %q", ErrNotFound, id))
	}

	result := pendingPlaceholder[*TwoDTexture]()
	s.ioWorker.Schedule(func() {
		texture, err := s.allocateTwoDTexture(resource)
		if err != nil {
			result.promise.Fail(fmt.Errorf("error loading twod texture %q: %w", id, err))
		} else {
			result.promise.Deliver(texture)
		}
	})
	s.namedTwoDTextures[id] = result
	return result
}

func (s *ResourceSet) OpenCubeTexture(id string) Placeholder[*CubeTexture] {
	if result, ok := s.findCubeTexture(id); ok {
		return result
	}

	resource := s.registry.ResourceByID(id)
	if resource == nil {
		return failedPlaceholder[*CubeTexture](fmt.Errorf("%w: %q", ErrNotFound, id))
	}

	result := pendingPlaceholder[*CubeTexture]()
	s.ioWorker.Schedule(func() {
		texture, err := s.allocateCubeTexture(resource)
		if err != nil {
			result.promise.Fail(fmt.Errorf("error loading cube texture %q: %w", id, err))
		} else {
			result.promise.Deliver(texture)
		}
	})
	s.namedCubeTextures[id] = result
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
		for _, placeholder := range s.namedTwoDTextures {
			if texture, err := placeholder.promise.Wait(); err == nil {
				s.releaseTwoDTexture(texture)
			}
		}
		s.namedTwoDTextures = nil
		for _, placeholder := range s.namedCubeTextures {
			if texture, err := placeholder.promise.Wait(); err == nil {
				s.releaseCubeTexture(texture)
			}
		}
		s.namedCubeTextures = nil
	}()
}

func (s *ResourceSet) findTwoDTexture(id string) (Placeholder[*TwoDTexture], bool) {
	if result, ok := s.namedTwoDTextures[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findTwoDTexture(id)
	}
	return Placeholder[*TwoDTexture]{}, false
}

func (s *ResourceSet) findCubeTexture(id string) (Placeholder[*CubeTexture], bool) {
	if result, ok := s.namedCubeTextures[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findCubeTexture(id)
	}
	return Placeholder[*CubeTexture]{}, false
}
