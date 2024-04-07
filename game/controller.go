package game

import (
	"time"

	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	newasset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/util/async"
)

func NewController(registry asset.Registry, shaders graphics.ShaderCollection, shaderBuilder graphics.ShaderBuilder) *Controller {
	return &Controller{
		registry:      registry,
		shaders:       shaders,
		shaderBuilder: shaderBuilder,
	}
}

var _ app.Controller = (*Controller)(nil)

type Controller struct {
	app.NopController

	registry      asset.Registry
	newRegistry   *newasset.Registry
	shaders       graphics.ShaderCollection
	shaderBuilder graphics.ShaderBuilder

	gfxEngine     *graphics.Engine
	ecsEngine     *ecs.Engine
	physicsEngine *physics.Engine

	window   app.Window
	ioWorker *async.Worker
	engine   *Engine

	viewport graphics.Viewport
}

func (c *Controller) SetNewRegistry(registry *newasset.Registry) {
	c.newRegistry = registry
}

func (c *Controller) Engine() *Engine {
	return c.engine
}

func (c *Controller) OnCreate(window app.Window) {
	c.window = window
	c.gfxEngine = graphics.NewEngine(window.RenderAPI(), c.shaders, c.shaderBuilder)
	c.ecsEngine = ecs.NewEngine()
	c.physicsEngine = physics.NewEngine(16 * time.Millisecond)

	c.ioWorker = async.NewWorker(16)
	go c.ioWorker.ProcessAll()

	c.engine = NewEngine(
		WithGFXWorker(c.gfxWorkerAdapter()),
		WithIOWorker(c.ioWorkerAdapter()),
		WithRegistry(c.registry),
		WithNewRegistry(c.newRegistry),
		WithGraphics(c.gfxEngine),
		WithECS(c.ecsEngine),
		WithPhysics(c.physicsEngine),
	)
	c.engine.Create()

	width, height := window.Size()
	c.OnResize(window, width, height)
}

func (c *Controller) OnDestroy(window app.Window) {
	c.engine.Destroy()
	c.ioWorker.Shutdown()
}

func (c *Controller) OnFramebufferResize(window app.Window, width, height int) {
	c.viewport = graphics.NewViewport(0, 0, uint32(width), uint32(height))
}

func (c *Controller) OnRender(window app.Window) {
	defer metric.BeginRegion("game").End()

	c.engine.Update()
	c.engine.Render(c.viewport)

	window.Invalidate() // force redraw
}

func (c *Controller) schedule(fn func()) {
	c.window.Schedule(fn)
}

func (c *Controller) gfxWorkerAdapter() Worker {
	return WorkerFunc(func(fn func() error) Operation {
		operation := NewOperation()
		c.schedule(func() {
			operation.Complete(fn())
		})
		return operation
	})
}

func (c *Controller) ioWorkerAdapter() Worker {
	return WorkerFunc(func(fn func() error) Operation {
		operation := NewOperation()
		c.ioWorker.Schedule(func() error {
			err := fn()
			operation.Complete(err)
			return err
		})
		return operation
	})
}
