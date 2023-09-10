package ui

import (
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/util/resource"
)

// InitFunc can be used to initialize the Window with the
// respective Element hierarchy.
type InitFunc func(window *Window)

// NewController creates a new app.Controller that integrates
// with the ui package to render a user interface.
func NewController(locator resource.ReadLocator, shaders ShaderCollection, initFn InitFunc) *Controller {
	return &Controller{
		locator: locator,
		shaders: shaders,
		initFn:  initFn,
	}
}

var _ app.Controller = (*Controller)(nil)

type Controller struct {
	locator resource.ReadLocator
	shaders ShaderCollection

	canvas  *Canvas
	fntFact *fontFactory
	resMan  *resourceManager
	initFn  InitFunc

	uiWindow        *Window
	uiWindowHandler WindowHandler
}

func (c *Controller) OnCreate(appWindow app.Window) {
	renderer := newCanvasRenderer(appWindow.RenderAPI(), c.shaders)

	c.canvas = newCanvas(renderer)
	c.fntFact = newFontFactory(appWindow.RenderAPI(), renderer)

	imgFact := newImageFactory(appWindow.RenderAPI())
	c.resMan = newResourceManager(c.locator, imgFact, c.fntFact)

	c.canvas.onCreate()
	c.fntFact.Init()

	c.uiWindow, c.uiWindowHandler = newWindow(appWindow, c.canvas, c.resMan)
	c.initFn(c.uiWindow)
}

func (c *Controller) OnResize(window app.Window, width, height int) {
	c.uiWindowHandler.OnResize(NewSize(width, height))
}

func (c *Controller) OnFramebufferResize(window app.Window, width, height int) {
	c.uiWindowHandler.OnFramebufferResize(NewSize(width, height))
}

func (c *Controller) OnKeyboardEvent(window app.Window, event app.KeyboardEvent) bool {
	return c.uiWindowHandler.OnKeyboardEvent(KeyboardEvent{
		Type:      event.Type,
		Code:      event.Code,
		Rune:      event.Rune,
		Modifiers: event.Modifiers,
	})
}

func (c *Controller) OnMouseEvent(window app.Window, event app.MouseEvent) bool {
	return c.uiWindowHandler.OnMouseEvent(MouseEvent{
		Index:    event.Index,
		Position: NewPosition(event.X, event.Y),
		Type:     event.Type,
		Button:   event.Button,
		Payload:  event.Payload,
		ScrollX:  event.ScrollX,
		ScrollY:  event.ScrollY,
	})
}

func (c *Controller) OnRender(window app.Window) {
	c.uiWindowHandler.OnRender()
}

func (c *Controller) OnCloseRequested(window app.Window) {
	c.uiWindowHandler.OnCloseRequested()
}

func (c *Controller) OnDestroy(window app.Window) {
	c.fntFact.Free()
	c.canvas.onDestroy()
}
