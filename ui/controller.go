package ui

import (
	"github.com/mokiat/lacking/app"
)

// InitFunc can be used to initialize the Window with the
// respective Element hierarchy.
type InitFunc func(window *Window)

// NewController creates a new app.Controller that integrates
// with the ui package to render a user interface.
func NewController(locator ResourceLocator, graphics Graphics, initFn InitFunc) app.Controller {
	return &controller{
		graphics: graphics,
		locator:  locator,
		initFn:   initFn,
	}
}

type controller struct {
	graphics Graphics
	locator  ResourceLocator
	initFn   InitFunc

	appWindow       app.Window
	uiWindow        *Window
	uiWindowHandler WindowHandler
}

func (c *controller) OnCreate(appWindow app.Window) {
	c.graphics.Create()

	c.appWindow = appWindow
	c.uiWindow, c.uiWindowHandler = NewWindow(appWindow, c.locator, c.graphics)
	c.initFn(c.uiWindow)
}

func (c *controller) OnResize(window app.Window, width, height int) {
	c.uiWindowHandler.OnResize(NewSize(width, height))
}

func (c *controller) OnFramebufferResize(window app.Window, width, height int) {
	c.uiWindowHandler.OnFramebufferResize(NewSize(width, height))
}

func (c *controller) OnKeyboardEvent(window app.Window, event app.KeyboardEvent) bool {
	return c.uiWindowHandler.OnKeyboardEvent(KeyboardEvent{
		Type:      event.Type,
		Code:      event.Code,
		Rune:      event.Rune,
		Modifiers: event.Modifiers,
	})
}

func (c *controller) OnMouseEvent(window app.Window, event app.MouseEvent) bool {
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

func (c *controller) OnRender(window app.Window) {
	c.uiWindowHandler.OnRender()
}

func (c *controller) OnCloseRequested(window app.Window) {
	c.uiWindowHandler.OnCloseRequested()
}

func (c *controller) OnDestroy(window app.Window) {
	c.graphics.Destroy()
}
