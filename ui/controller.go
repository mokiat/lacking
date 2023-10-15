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
	app.NopController

	locator resource.ReadLocator
	shaders ShaderCollection

	canvas  *Canvas
	fntFact *fontFactory
	resMan  *resourceManager
	initFn  InitFunc

	uiWindow        *Window
	uiWindowHandler WindowHandler

	modifierLeftControl  bool
	modifierRightControl bool
	modifierLeftShift    bool
	modifierRightShift   bool
	modifierLeftAlt      bool
	modifierRightAlt     bool
	modifierLeftSuper    bool
	modifierRightSuper   bool
	modifierCapsLock     bool
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
	c.trackModifiers(event)
	return c.uiWindowHandler.OnKeyboardEvent(KeyboardEvent{
		Action:    KeyboardAction(event.Action),
		Code:      KeyCode(event.Code),
		Rune:      event.Character,
		Modifiers: c.buildModifierSet(),
	})
}

func (c *Controller) OnMouseEvent(window app.Window, event app.MouseEvent) bool {
	return c.uiWindowHandler.OnMouseEvent(MouseEvent{
		Index:     event.Index,
		Action:    MouseAction(event.Action),
		Button:    MouseButton(event.Button),
		X:         event.X,
		Y:         event.Y,
		ScrollX:   int(event.ScrollX * 100.0),
		ScrollY:   int(event.ScrollY * 100.0),
		Modifiers: c.buildModifierSet(),
		Payload:   event.Payload,
	})
}

func (c *Controller) OnClipboardEvent(window app.Window, event app.ClipboardEvent) bool {
	return c.uiWindowHandler.OnClipboardEvent(ClipboardEvent{
		Action: ClipboardActionPaste,
		Text:   event.Text,
	})
}

func (c *Controller) OnRender(window app.Window) {
	c.uiWindowHandler.OnRender()
}

func (c *Controller) OnCloseRequested(window app.Window) bool {
	return c.uiWindowHandler.OnCloseRequested()
}

func (c *Controller) OnDestroy(window app.Window) {
	c.fntFact.Free()
	c.canvas.onDestroy()
}

func (c *Controller) trackModifiers(event app.KeyboardEvent) {
	var pressed bool
	switch event.Action {
	case app.KeyboardActionUp:
		pressed = false
	case app.KeyboardActionDown:
		pressed = true
	default:
		return // does not affect modifiers
	}
	switch event.Code {
	case app.KeyCodeLeftControl:
		c.modifierLeftControl = pressed
	case app.KeyCodeRightControl:
		c.modifierRightControl = pressed
	case app.KeyCodeLeftShift:
		c.modifierLeftShift = pressed
	case app.KeyCodeRightShift:
		c.modifierRightShift = pressed
	case app.KeyCodeLeftAlt:
		c.modifierLeftAlt = pressed
	case app.KeyCodeRightAlt:
		c.modifierRightAlt = pressed
	case app.KeyCodeLeftSuper:
		c.modifierLeftSuper = pressed
	case app.KeyCodeRightSuper:
		c.modifierRightSuper = pressed
	case app.KeyCodeCaps:
		c.modifierCapsLock = pressed
	}
}

func (c *Controller) buildModifierSet() KeyModifierSet {
	var result KeyModifierSet
	if c.modifierLeftControl || c.modifierRightControl {
		result |= KeyModifierSet(KeyModifierControl)
	}
	if c.modifierLeftShift || c.modifierRightShift {
		result |= KeyModifierSet(KeyModifierShift)
	}
	if c.modifierLeftAlt || c.modifierRightAlt {
		result |= KeyModifierSet(KeyModifierAlt)
	}
	if c.modifierLeftSuper || c.modifierRightSuper {
		result |= KeyModifierSet(KeyModifierSuper)
	}
	if c.modifierCapsLock {
		result |= KeyModifierSet(KeyModifierCapsLock)
	}
	return result
}
