package ui

import (
	"fmt"

	"github.com/mokiat/lacking/app"
)

// NewWindow creates a new Window instance that integrates
// with the specified app.Window.
func NewWindow(appWindow app.Window, locator ResourceLocator, graphics Graphics) (*Window, WindowHandler) {
	window := &Window{
		Window:   appWindow,
		graphics: graphics,
	}
	window.context = newContext(window, locator, graphics)
	window.root = newElement(window.context)
	window.root.SetLayout(NewFillLayout())
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
	app.Window
	graphics Graphics
	context  *Context

	size           Size
	root           *Element
	pointedElement *Element
}

// Size returns the content area of this Window.
func (w *Window) Size() Size {
	return NewSize(w.Window.Size())
}

// SetSize changes the content area of this Window
// to the specified size.
func (w *Window) SetSize(size Size) {
	w.Window.SetSize(size.Width, size.Height)
}

// Context returns the Context for this Window.
//
// Note that anything allocated through this Context
// will not be released until this Window is closed.
func (w *Window) Context() *Context {
	return w.context
}

// Root returns the Element that represents this Window.
func (w *Window) Root() *Element {
	return w.root
}

// FindElementByID looks through the Element hierarchy tree for an Element
// with the specified ID.
func (w *Window) FindElementByID(id string) (*Element, bool) {
	return w.dfsElementByID(w.root, id)
}

// GetElementByID looks through the Element hierarchy tree for an Element
// with the specified ID. Unlike FindElementByID, this method panics if
// such an Element cannot be found.
func (w *Window) GetElementByID(id string) *Element {
	element, found := w.FindElementByID(id)
	if !found {
		panic(fmt.Errorf("element with id %q not found", id))
	}
	return element
}

func (w *Window) dfsElementByID(current *Element, id string) (*Element, bool) {
	if current == nil {
		return nil, false
	}
	if current.id == id {
		return current, true
	}
	for child := current.firstChild; child != nil; child = child.rightSibling {
		if result, found := w.dfsElementByID(child, id); found {
			return result, true
		}
	}
	return nil, false
}

type windowHandler struct {
	*Window
}

func (w *windowHandler) OnResize(size Size) {
	w.graphics.Resize(size)
	w.size = size

	if w.root != nil {
		w.root.SetBounds(Bounds{
			Position: NewPosition(0, 0),
			Size:     size,
		})
	}
}

func (w *windowHandler) OnFramebufferResize(size Size) {
	w.graphics.ResizeFramebuffer(size)
}

func (w *windowHandler) OnKeyboardEvent(event KeyboardEvent) bool {
	// TODO
	return false
}

func (w *windowHandler) OnMouseEvent(event MouseEvent) bool {
	if w.root != nil {
		return w.processMouseEvent(w.root, event)
	}
	return false
}

func (w *windowHandler) OnRender() {
	w.graphics.Begin()

	if w.root != nil {
		w.renderElement(w.root, w.graphics.Canvas(), Bounds{
			Position: NewPosition(0, 0),
			Size:     w.size,
		})
	}

	w.graphics.End()
}

func (w *windowHandler) OnCloseRequested() {
	w.Close()
}

func (w *windowHandler) processMouseEvent(element *Element, event MouseEvent) bool {
	if !element.enabled || !element.visible {
		return false
	}

	// Check if any of the children (from top to bottom) can process the event.
	for childElement := element.lastChild; childElement != nil; childElement = childElement.leftSibling {
		if childBounds := childElement.Bounds(); childBounds.Contains(event.Position) {
			translatedEvent := event
			translatedEvent.Position = event.Position.Translate(-childBounds.X, -childBounds.Y)
			if w.processMouseEvent(childElement, translatedEvent) {
				return true
			}
		}
	}

	// Check if we need to change mouse ownership.
	if element != w.pointedElement {
		if w.pointedElement != nil {
			w.pointedElement.onMouseEvent(MouseEvent{
				Index:    event.Index,
				Position: event.Position,
				Type:     MouseEventTypeLeave,
				Button:   event.Button,
			})
		}

		element.onMouseEvent(MouseEvent{
			Index:    event.Index,
			Position: event.Position,
			Type:     MouseEventTypeEnter,
			Button:   event.Button,
		})

		w.pointedElement = element
	}

	// Let the current element handle the event.
	return element.onMouseEvent(event)
}

func (w *Window) renderElement(element *Element, canvas Canvas, dirtyRegion Bounds) {
	dirtyRegion = dirtyRegion.Intersect(element.bounds)
	if dirtyRegion.Empty() {
		return
	}

	canvas.Push()
	canvas.Clip(element.bounds)
	canvas.Translate(element.bounds.Position)
	element.onRender(canvas)
	if contentBounds := element.ContentBounds(); !contentBounds.Empty() {
		canvas.Clip(contentBounds)
		for child := element.firstChild; child != nil; child = child.rightSibling {
			w.renderElement(child, canvas, dirtyRegion.Translate(element.bounds.Position.Inverse()))
		}
	}
	canvas.Pop()
}
