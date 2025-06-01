package game

import (
	"errors"
	"fmt"

	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/storage/chunked"
	"github.com/mokiat/lacking/util/async"
)

// ErrNotFound indicates that a resource was not found.
var ErrNotFound = errors.New("resource not found")

func newResourceSet(parent *ResourceSet, engine *Engine) *ResourceSet {
	return &ResourceSet{
		parent: parent,

		engine:    engine,
		storage:   engine.Storage(),
		renderAPI: engine.Graphics().API(),

		ioWorker:  engine.ioWorker,
		gfxWorker: engine.gfxWorker,

		namedModels: make(map[string]async.Promise[*ModelTemplate]),
	}
}

// ResourceSet is a collection of resources that are managed together.
type ResourceSet struct {
	parent *ResourceSet

	engine    *Engine
	storage   chunked.Storage
	renderAPI render.API

	ioWorker  Worker
	gfxWorker Worker

	namedModels map[string]async.Promise[*ModelTemplate]
}

// CreateResourceSet creates a new ResourceSet that inherits resources from
// the current one. Opening a resource in the new resource set will first
// check if the current one has it.
func (s *ResourceSet) CreateResourceSet() *ResourceSet {
	return newResourceSet(s, s.engine)
}

// OpenModel loads a model definition by its path.
func (s *ResourceSet) OpenModel(path string) async.Promise[*ModelTemplate] {
	if result, ok := s.findModel(path); ok {
		return result
	}
	resource := chunked.NewAsset(s.storage, path)

	result := async.NewPromise[*ModelTemplate]()
	s.engine.RunAsync(func(asyncEngine *AsyncEngine) error {
		model, err := s.loadModel(asyncEngine, resource)
		if err != nil {
			result.Fail(fmt.Errorf("error loading model %q: %w", path, err))
		} else {
			result.Deliver(model)
		}
		return err
	})
	s.namedModels[path] = result
	return result
}

// Delete schedules all resources managed by this ResourceSet for deletion.
// After this method returns, the resources are not guaranteed to have been
// released.
func (s *ResourceSet) Delete() {
	for _, promise := range s.namedModels {
		go func() {
			if model, err := promise.Wait(); err == nil {
				s.freeModel(model)
			}
		}()
	}
	clear(s.namedModels)
}

func (s *ResourceSet) findModel(path string) (async.Promise[*ModelTemplate], bool) {
	if result, ok := s.namedModels[path]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findModel(path)
	}
	return async.Promise[*ModelTemplate]{}, false
}

// AssetLoader represents an async loading process in the scope of a given
// ResourceSet.
type AssetLoader struct {
	resourceSet *ResourceSet
	asyncEngine *AsyncEngine
}

func (l *AssetLoader) ScheduleIO(cb func() error) async.Operation {
	return l.asyncEngine.ScheduleIO(cb)
}

func (l *AssetLoader) ScheduleMain(cb func(engine *Engine) error) async.Operation {
	return l.asyncEngine.ScheduleMain(cb)
}
