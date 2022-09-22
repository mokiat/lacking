package game

import (
	"time"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/util/datastruct"
	"github.com/mokiat/lacking/util/metrics"
)

type EngineOption func(e *Engine)

func WithRegistry(registry asset.Registry) EngineOption {
	return func(e *Engine) {
		e.registry = registry
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

func WithECS(ecsEngine *ecs.Engine) EngineOption {
	return func(e *Engine) {
		e.ecsEngine = ecsEngine
	}
}

func NewEngine(opts ...EngineOption) *Engine {
	result := &Engine{
		preUpdateSubscriptions:  datastruct.NewList[*UpdateSubscription](),
		postUpdateSubscriptions: datastruct.NewList[*UpdateSubscription](),

		lastTick: time.Now(),
	}
	for _, opt := range opts {
		opt(result)
	}
	return result
}

type Engine struct {
	registry      asset.Registry
	ioWorker      Worker
	gfxWorker     Worker
	physicsEngine *physics.Engine
	gfxEngine     *graphics.Engine
	ecsEngine     *ecs.Engine

	preUpdateSubscriptions  *datastruct.List[*UpdateSubscription]
	postUpdateSubscriptions *datastruct.List[*UpdateSubscription]

	activeScene *Scene
	lastTick    time.Time
}

func (e *Engine) Create() {
	e.gfxEngine.Create()
	e.ResetDeltaTime()
}

func (e *Engine) Destroy() {
	e.preUpdateSubscriptions.Clear()
	e.postUpdateSubscriptions.Clear()
	e.gfxEngine.Destroy()
	// TODO: Release all scenes and all resource sets
}

func (e *Engine) Registry() asset.Registry {
	return e.registry
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

func (e *Engine) ECS() *ecs.Engine {
	return e.ecsEngine
}

func (e *Engine) SubscribePreUpdate(callback UpdateCallback) *UpdateSubscription {
	result := &UpdateSubscription{
		list:     e.preUpdateSubscriptions,
		callback: callback,
	}
	e.preUpdateSubscriptions.Add(result)
	return result
}

func (e *Engine) SubscribePostUpdate(callback UpdateCallback) *UpdateSubscription {
	result := &UpdateSubscription{
		list:     e.postUpdateSubscriptions,
		callback: callback,
	}
	e.postUpdateSubscriptions.Add(result)
	return result
}

func (e *Engine) ActiveScene() *Scene {
	return e.activeScene
}

func (e *Engine) SetActiveScene(scene *Scene) {
	e.activeScene = scene
}

func (e *Engine) CreateResourceSet() *ResourceSet {
	return newResourceSet(nil, e)
}

func (e *Engine) CreateScene() *Scene {
	physicsScene := e.physicsEngine.CreateScene(0.015)
	gfxScene := e.gfxEngine.CreateScene()
	ecsScene := e.ecsEngine.CreateScene()
	resourceSet := e.CreateResourceSet()
	result := newScene(resourceSet, physicsScene, gfxScene, ecsScene)
	if e.activeScene == nil {
		e.activeScene = result
	}
	return result
}

func (e *Engine) CreateAnimationDefinition(info AnimationDefinitionInfo) *AnimationDefinition {
	return &AnimationDefinition{
		name:      info.Name,
		startTime: info.StartTime,
		endTime:   info.EndTime,
		bindings:  info.Bindings,
	}
}

func (e *Engine) ResetDeltaTime() {
	e.lastTick = time.Now()
}

func (e *Engine) Update() {
	defer metrics.BeginSpan("update").End()

	currentTime := time.Now()
	elapsedSeconds := currentTime.Sub(e.lastTick).Seconds()
	e.lastTick = currentTime

	if e.activeScene != nil {
		e.preUpdateSubscriptions.Each(func(sub *UpdateSubscription) {
			sub.callback(e, e.activeScene, elapsedSeconds)
		})
		e.activeScene.Update(elapsedSeconds)
		e.postUpdateSubscriptions.Each(func(sub *UpdateSubscription) {
			sub.callback(e, e.activeScene, elapsedSeconds)
		})
	}
}

func (e *Engine) Render(viewport graphics.Viewport) {
	defer metrics.BeginSpan("render").End()

	if e.activeScene != nil {
		e.activeScene.Graphics().Render(viewport)
	}
}
