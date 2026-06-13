package game

import (
	"time"

	"github.com/mokiat/lacking/core/resource"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/async"
)

type EngineOption func(e *Engine)

func WithStore(registry resource.Store) EngineOption {
	return func(e *Engine) {
		e.store = registry
	}
}

func WithIOWorker(worker Worker) EngineOption {
	return func(e *Engine) {
		e.ioWorker = worker
	}
}

func WithGFXWorker(worker Worker) EngineOption {
	return func(e *Engine) {
		e.gfxWorker = worker
	}
}

func WithPhysics(physicsEngine *physics.Engine) EngineOption {
	return func(e *Engine) {
		e.physicsEngine = physicsEngine
	}
}

func WithGraphics(gfxEngine *graphics.Engine) EngineOption {
	return func(e *Engine) {
		e.gfxEngine = gfxEngine
	}
}

func NewEngine(opts ...EngineOption) *Engine {
	result := &Engine{
		lastTick: time.Now(),
	}
	for _, opt := range opts {
		opt(result)
	}
	result.registry = newResourceRegistry(result, result.store)
	result.registry.RegisterResourceLoader(newModelResourceLoader())
	return result
}

type Engine struct {
	store         resource.Store
	ioWorker      Worker
	gfxWorker     Worker
	physicsEngine *physics.Engine
	gfxEngine     *graphics.Engine

	registry *resourceRegistry

	activeScene *Scene
	lastTick    time.Time
}

func (e *Engine) Create() {
	e.gfxEngine.Create()
	e.ResetDeltaTime()
}

func (e *Engine) Destroy() {
	e.gfxEngine.Destroy()
	// TODO: Release all scenes and all resource sets
}

func (e *Engine) Storage() resource.Store {
	return e.store
}

func (e *Engine) IOWorker() Worker {
	return e.ioWorker
}

func (e *Engine) GFXWorker() Worker {
	return e.gfxWorker
}

func (e *Engine) Physics() *physics.Engine {
	return e.physicsEngine
}

func (e *Engine) Graphics() *graphics.Engine {
	return e.gfxEngine
}

func (e *Engine) ActiveScene() *Scene {
	return e.activeScene
}

func (e *Engine) SetActiveScene(scene *Scene) {
	e.activeScene = scene
}

// CreateResourceSet creates a new ResourceSet that can be used to manage
// resources together.
func (e *Engine) CreateResourceSet() *ResourceSet {
	return newResourceSet(e, e.registry)
}

func (e *Engine) RegisterResourceLoader(resourceLoader ResourceLoader[any]) {
	e.registry.RegisterResourceLoader(resourceLoader)
}

func (e *Engine) UnregisterResourceLoader(resourceLoader ResourceLoader[any]) {
	e.registry.UnregisterResourceLoader(resourceLoader)
}

func (e *Engine) CreateScene(info SceneInfo) *Scene {
	result := newScene(e, info)
	if e.activeScene == nil {
		e.activeScene = result
	}
	return result
}

func (e *Engine) ResetDeltaTime() {
	e.lastTick = time.Now()
}

func (e *Engine) ScheduleIO(cb func() error) async.Operation {
	result := async.NewOperation()
	e.ioWorker.Schedule(func() {
		if err := cb(); err == nil {
			result.Pass()
		} else {
			result.Fail(err)
		}
	})
	return result
}

func (e *Engine) ScheduleMain(cb func() error) async.Operation {
	result := async.NewOperation()
	e.gfxWorker.Schedule(func() {
		if err := cb(); err == nil {
			result.Pass()
		} else {
			result.Fail(err)
		}
	})
	return result
}

func (e *Engine) Update() {
	e.gfxEngine.Debug().Reset()

	currentTime := time.Now()
	elapsedTime := currentTime.Sub(e.lastTick)
	e.lastTick = currentTime

	if e.activeScene != nil {
		e.activeScene.Update(elapsedTime)
	}
}

func (e *Engine) Render(framebuffer render.Framebuffer, viewport graphics.Viewport) {
	if e.activeScene != nil {
		e.activeScene.Render(framebuffer, viewport)
	}
}
