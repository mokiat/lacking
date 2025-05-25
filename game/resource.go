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

		namedModels: make(map[string]async.Promise[*ModelDefinition]),
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

	namedModels map[string]async.Promise[*ModelDefinition]
}

// CreateResourceSet creates a new ResourceSet that inherits resources from
// the current one. Opening a resource in the new resource set will first
// check if the current one has it.
func (s *ResourceSet) CreateResourceSet() *ResourceSet {
	return newResourceSet(s, s.engine)
}

// OpenModelByID loads a model definition by its ID.
func (s *ResourceSet) OpenModelByID(id string) async.Promise[*ModelDefinition] {
	if result, ok := s.findModel(id); ok {
		return result
	}

	resource := chunked.NewAsset(s.storage, id)

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

// OpenModelByName loads a model definition by its name.
func (s *ResourceSet) OpenModelByName(name string) async.Promise[*ModelDefinition] {
	return s.OpenModelByID(name)
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

func (s *ResourceSet) findModel(id string) (async.Promise[*ModelDefinition], bool) {
	if result, ok := s.namedModels[id]; ok {
		return result, true
	}
	if s.parent != nil {
		return s.parent.findModel(id)
	}
	return async.Promise[*ModelDefinition]{}, false
}
