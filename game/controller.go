package game

import (
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/util/async"
)

// NewController creates a new game controller that manages the lifecycle
// of a game engine. The controller will use the provided asset registry
// to load and manage assets. The provided shader collection will be used
// to render the game. The provided shader builder will be used to create
// new shaders when needed.
func NewController(registry *asset.Registry, shaders graphics.ShaderCollection, shaderBuilder graphics.ShaderBuilder) *Controller {
	return &Controller{
		registry:      registry,
		shaders:       shaders,
		shaderBuilder: shaderBuilder,
	}
}

var _ app.Controller = (*Controller)(nil)

// Controller is an implementation of the app.Controller interface which
// initializes a game engine and manages its lifecycle. Furthermore, it
// ensures that the game engine is updated and rendered on each frame.
type Controller struct {
	app.NopController

	registry      *asset.Registry
	shaders       graphics.ShaderCollection
	shaderBuilder graphics.ShaderBuilder

	gfxOptions []graphics.Option
	gfxEngine  *graphics.Engine

	ecsOptions []ecs.Option
	ecsEngine  *ecs.Engine

	physicsOptions []physics.Option
	physicsEngine  *physics.Engine

	window   app.Window
	ioWorker *async.Worker
	engine   *Engine

	viewport graphics.Viewport
}

// Registry returns the asset registry to be used by the game.
func (c *Controller) Registry() *asset.Registry {
	return c.registry
}

// UseGraphicsOptions allows to specify options that will be used
// when initializing the graphics engine. This method should be
// called before the controller is initialized by the app framework.
func (c *Controller) UseGraphicsOptions(opts ...graphics.Option) {
	c.gfxOptions = opts
}

// UseECSOptions allows to specify options that will be used
// when initializing the ECS engine. This method should be
// called before the controller is initialized by the app framework.
func (c *Controller) UseECSOptions(opts ...ecs.Option) {
	c.ecsOptions = opts
}

// UsePhysicsOptions allows to specify options that will be used
// when initializing the physics engine. This method should be
// called before the controller is initialized by the app framework.
func (c *Controller) UsePhysicsOptions(opts ...physics.Option) {
	c.physicsOptions = opts
}

// Engine returns the game engine that is managed by the controller.
//
// This method should only be called after the controller has been
// initialized by the app framework.
func (c *Controller) Engine() *Engine {
	return c.engine
}

func (c *Controller) OnCreate(window app.Window) {
	c.window = window
	c.gfxEngine = graphics.NewEngine(window.RenderAPI(), c.shaders, c.shaderBuilder, c.gfxOptions...)
	c.ecsEngine = ecs.NewEngine(c.ecsOptions...)
	c.physicsEngine = physics.NewEngine(c.physicsOptions...)

	c.ioWorker = async.NewWorker(4)
	go c.ioWorker.ProcessAll()

	c.engine = NewEngine(
		WithGFXWorker(window),
		WithIOWorker(c.ioWorker),
		WithRegistry(c.registry),
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
	c.engine.Render(window.RenderAPI().DefaultFramebuffer(), c.viewport)

	window.Invalidate() // force redraw
}
