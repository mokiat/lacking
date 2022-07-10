package game

import (
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
)

type EngineOption func(e *Engine)

func WithRegistry(registry *asset.Registry) EngineOption {
	return func(e *Engine) {
		e.registry = registry
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
	result := &Engine{}
	for _, opt := range opts {
		opt(result)
	}
	return result
}

type Engine struct {
	registry      *asset.Registry
	physicsEngine *physics.Engine
	gfxEngine     *graphics.Engine
	ecsEngine     *ecs.Engine
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

func (e *Engine) CreateScene() *Scene {
	physicsScene := e.physicsEngine.CreateScene(0.015)
	gfxScene := e.gfxEngine.CreateScene()
	ecsScene := e.ecsEngine.CreateScene()
	return newScene(physicsScene, gfxScene, ecsScene)
}

// func (e *Engine) OpenTwoDTexture(resourceSet *ResourceSet, id string) async.Promise[*graphics.TwoDTexture] {
// 	return nil // TODO
// }

// func (e *Engine) OpenCubeTexture(resourceSet *ResourceSet, id string) async.Promise[*graphics.CubeTexture] {
// 	return nil // TODO
// }

// func (e *Engine) OpenModel(resourceSet *ResourceSet, id string) async.Promise[*]
