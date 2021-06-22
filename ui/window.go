package ui

import (
	"fmt"

	"github.com/mokiat/lacking/app"
)

// NewWindow creates a new Window instance that integrates
// with the specified app.Window.
func NewWindow(appWindow app.Window, locator ResourceLocator, graphics Graphics) (*Window, WindowHandler) {
	window := &Window{
		appWindow: appWindow,
		graphics:  graphics,
		locator:   locator,
	}
	handler := &windowHandler{
		Window: window,
	}
	return window, handler
}

// WindowHandler is an interface that is used by
// the framework to communicate with Window implementations
// critical events.
type WindowHandler interface {

	// OnResize is called whenever the native window has
	// been resized.
	OnResize(size Size)

	// OnFramebufferResize is called whenever the native window's
	// framebuffer has been resized.
	OnFramebufferResize(size Size)

	// OnKeyboardEvent is called whenever a native key event
	// has been registered.
	OnKeyboardEvent(event KeyboardEvent) bool

	// OnMouseEvent is called whenever a native mouse event
	// has been registered.
	OnMouseEvent(event MouseEvent) bool

	// OnRender is called whenever the Window should redraw
	// itself.
	OnRender()

	// OnCloseRequested is called whenever the end-user has
	// indicated that they would like to close the appplication.
	// (e.g. using the close button on the application)
	OnCloseRequested()
}

// Window represents an application window.
type Window struct {
	appWindow app.Window
	graphics  Graphics
	locator   ResourceLocator

	size            Size
	activeViewLayer *viewLayer
}

// Title returns the title of this Window.
func (w *Window) Title() string {
	return w.appWindow.Title()
}

// SetTitle changes the title of the Window to the
// specified value.
func (w *Window) SetTitle(title string) {
	w.appWindow.SetTitle(title)
}

// Size returns the content area of this Window.
func (w *Window) Size() Size {
	return NewSize(w.appWindow.Size())
}

// SetSize changes the content area of this Window
// to the specified size.
func (w *Window) SetSize(size Size) {
	w.appWindow.SetSize(size.Width, size.Height)
}

// OpenView opens a new View which replaces the current
// one according to the specified ViewMode.
//
// The preparation of the View is performed by the specified
// ViewType.
func (w *Window) OpenView(mode ViewMode, vType ViewType) error {
	context := newContext(w, w.locator, w.graphics)
	view := newView(w, context)
	if err := vType.SetupView(view); err != nil {
		return fmt.Errorf("failed to setup view: %w", err)
	}

	if w.activeViewLayer != nil {
		switch {
		case mode.Is(ViewModeReplace):
			w.activeViewLayer.view.onHide()
			w.activeViewLayer.view.onDestroy()
			w.activeViewLayer = w.activeViewLayer.parent
		case mode.Is(ViewModeCover):
			w.activeViewLayer.view.onHide()
			w.activeViewLayer.visible = false
		}
	}

	w.activeViewLayer = &viewLayer{
		parent:  w.activeViewLayer,
		mode:    mode,
		view:    view,
		visible: true,
	}
	w.activeViewLayer.view.onCreate()
	w.activeViewLayer.view.onShow()
	w.activeViewLayer.view.onResize(w.size)

	w.Invalidate()
	return nil
}

// Invalidate causes this Window to be redrawn when possible.
func (w *Window) Invalidate() {
	w.appWindow.Invalidate()
}

// Destroy closes this Window and releases all resources
// allocated to it.
func (w *Window) Destroy() {
	w.appWindow.Close()
}

type windowHandler struct {
	*Window
}

func (w *windowHandler) OnResize(size Size) {
	w.graphics.Resize(size)
	w.size = size

	for layer := w.activeViewLayer; layer != nil && layer.visible; layer = layer.parent {
		layer.view.onResize(size)
	}
}

func (w *windowHandler) OnFramebufferResize(size Size) {
	w.graphics.ResizeFramebuffer(size)
}

func (w *windowHandler) OnKeyboardEvent(event KeyboardEvent) bool {
	if w.activeViewLayer != nil {
		return w.activeViewLayer.view.onKeyboardEvent(event)
	}
	return false
}

func (w *windowHandler) OnMouseEvent(event MouseEvent) bool {
	if w.activeViewLayer != nil {
		return w.activeViewLayer.view.onMouseEvent(event)
	}
	return false
}

func (w *windowHandler) OnRender() {
	w.graphics.Begin()
	if w.activeViewLayer != nil {
		w.renderLayer(w.activeViewLayer)
	}
	w.graphics.End()
}

func (w *windowHandler) OnCloseRequested() {
	w.Destroy()
}

func (w *windowHandler) renderLayer(layer *viewLayer) {
	if layer.parent != nil && layer.parent.visible {
		w.renderLayer(layer.parent)
	}

	layer.view.onRender(w.graphics.Canvas(), Bounds{
		Position: NewPosition(0, 0),
		Size:     w.size,
	})
}

type viewLayer struct {
	parent  *viewLayer
	mode    ViewMode
	view    *View
	visible bool
}
