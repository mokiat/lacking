package ui

import (
	"fmt"

	"github.com/mokiat/lacking/app"
)

// NewController creates a new app.Controller that integrates
// with the ui package to render a user interface.
func NewController(locator ResourceLocator, graphics Graphics, initView ViewType) app.Controller {
	return &controller{
		graphics: graphics,
		locator:  locator,
		initView: initView,
	}
}

type controller struct {
	graphics Graphics
	locator  ResourceLocator
	initView ViewType

	appWindow       app.Window
	uiWindow        *Window
	uiWindowHandler WindowHandler
}

func (c *controller) OnCreate(appWindow app.Window) {
	c.graphics.Create()

	c.appWindow = appWindow
	c.uiWindow, c.uiWindowHandler = NewWindow(appWindow, c.locator, c.graphics)
	if err := c.uiWindow.OpenView(ViewModeNone, c.initView); err != nil {
		panic(fmt.Errorf("failed to open view: %w", err))
	}
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
