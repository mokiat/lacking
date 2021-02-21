package resource

import (
	"fmt"
	"sync"

	"github.com/mokiat/lacking/async"
)

// NOTE: Loading of resources always happens through resource sets.
// Reference counting happens through resource sets as well. The only
// way to release resources is to release a ResourceSet, not individual
// resources.

type Allocator interface {
	Allocate(set *Set) (interface{}, error)
}

type AllocatorFunc func(set *Set) (interface{}, error)

func (f AllocatorFunc) Allocate(set *Set) (interface{}, error) {
	return f(set)
}

type Releaser interface {
	Release(value interface{}) error
}

type ReleaserFunc func(value interface{}) error

func (f ReleaserFunc) Release(value interface{}) error {
	return f(value)
}

func NewRegistry(locator Locator, gfxWorker *async.Worker) *Registry {
	registry := &Registry{
		programOperator:     NewProgramOperator(locator, gfxWorker),
		twodTextureOperator: NewTwoDTextureOperator(locator, gfxWorker),
		cubeTextureOperator: NewCubeTextureOperator(locator, gfxWorker),
		modelOperator:       NewModelOperator(locator, gfxWorker),
		levelOperator:       NewLevelOperator(locator, gfxWorker),
		shaderOperator:      NewShaderOperator(gfxWorker),

		resources: make(map[string]*resourceEntry),
	}
	return registry
}

type Registry struct {
	programOperator     *ProgramOperator
	twodTextureOperator *TwoDTextureOperator
	cubeTextureOperator *CubeTextureOperator
	modelOperator       *ModelOperator
	levelOperator       *LevelOperator
	shaderOperator      *ShaderOperator

	resourceMU sync.Mutex
	resources  map[string]*resourceEntry
}

func (r *Registry) allocate(set *Set, id string, allocator Allocator, releaser Releaser, inject func(value interface{})) async.Eventual {
	r.resourceMU.Lock()
	defer r.resourceMU.Unlock()

	eventual, eventualDone := async.NewEventual()

	entry, ok := r.resources[id]
	if ok {
		entry.count++

		go func() {
			err := entry.eventual.Wait()
			if err == nil {
				inject(entry.value)
			}
			eventualDone(err)
		}()

	} else {
		entry := &resourceEntry{
			releaser: releaser,
			count:    1,
			eventual: eventual,
		}
		r.resources[id] = entry

		go func() {
			value, err := allocator.Allocate(set)
			if err == nil {
				entry.value = value
				inject(value)
			}
			eventualDone(err)
		}()
	}

	return eventual
}

func (r *Registry) release(id string, count int) async.Eventual {
	r.resourceMU.Lock()
	defer r.resourceMU.Unlock()

	entry, ok := r.resources[id]
	if !ok {
		panic(fmt.Errorf("releasing a resource that is already released"))
	}

	entry.count -= count
	if entry.count > 0 {
		return async.ImmediateEventual(nil)
	}

	delete(r.resources, id)

	eventual, eventualDone := async.NewEventual()
	go func() {
		// it could still be loading
		if err := entry.eventual.Wait(); err != nil {
			eventualDone(fmt.Errorf("failed to release entry as it was not created successfully: %w", err))
			return
		}

		eventualDone(entry.releaser.Release(entry.value))
	}()
	return eventual
}

type resourceEntry struct {
	releaser Releaser
	count    int
	eventual async.Eventual
	value    interface{}
}
