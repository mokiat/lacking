package ui

import (
	"fmt"

	"github.com/mokiat/lacking/app"
)

// newWindow creates a new Window instance that integrates
// with the specified app.Window.
func newWindow(appWindow app.Window, canvas *Canvas, resMan *resourceManager) (*Window, WindowHandler) {
	window := &Window{
		Window: appWindow,
		canvas: canvas,
	}
	window.context = newContext(nil, window, resMan)
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

	// TOOD: Remove this interface and make it all internal

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
	context *Context
	canvas  *Canvas

	size           Size
	root           *Element
	focusedElement *Element

	oldMousePosition Position
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

// IsElementFocused returns whether the specified element is the currently
// focused Element.
func (w *Window) IsElementFocused(element *Element) bool {
	return w.focusedElement == element
}

// DiscardFocus removes the focus from any Element.
func (w *Window) DiscardFocus() {
	w.focusedElement = nil
	w.Invalidate()
}

type windowHandler struct {
	*Window
}

func (w *windowHandler) OnResize(size Size) {
	w.canvas.onResize(size)
	w.size = size
	w.root.SetBounds(Bounds{
		Position: NewPosition(0, 0),
		Size:     size,
	})
}

func (w *windowHandler) OnFramebufferResize(size Size) {
	w.canvas.onResizeFramebuffer(size)
}

func (w *windowHandler) OnKeyboardEvent(event KeyboardEvent) bool {
	current := w.focusedElement
	for current != nil {
		if current.focusable && current.onKeyboardEvent(event) {
			return true
		}
		current = current.parent
	}
	return false
}

func (w *windowHandler) OnMouseEvent(event MouseEvent) bool {
	// TODO: Use a better algorithm. This one does not handle resize.
	w.processMouseLeave(w.root, event.Position, w.oldMousePosition)
	w.processMouseEnter(w.root, event.Position, w.oldMousePosition)
	w.oldMousePosition = event.Position
	if event.Type == MouseEventTypeDown {
		oldFocusedElement := w.focusedElement
		w.processFocusChange(w.root, event.Position)
		if w.focusedElement != oldFocusedElement {
			w.Invalidate()
		}
	}
	return w.processMouseEvent(w.root, event)
}

func (w *windowHandler) OnRender() {
	w.canvas.onBegin()
	w.canvas.Clip(Bounds{
		Position: NewPosition(0, 0),
		Size:     w.size,
	})
	w.renderElement(w.root, w.canvas, Bounds{
		Position: NewPosition(0, 0),
		Size:     w.size,
	})
	w.canvas.onEnd()
}

func (w *windowHandler) OnCloseRequested() {
	w.Close()
}

func (w *windowHandler) processFocusChange(element *Element, position Position) {
	if !element.enabled || !element.visible {
		return
	}

	bounds := element.Bounds()
	if !bounds.Contains(position) {
		return
	}

	if element.focusable {
		w.focusedElement = element
	}

	relativePosition := position.Translate(-bounds.X, -bounds.Y)
	for childElement := element.lastChild; childElement != nil; childElement = childElement.leftSibling {
		w.processFocusChange(childElement, relativePosition)
	}
}

func (w *windowHandler) processMouseLeave(element *Element, newPosition, oldPosition Position) {
	if !element.enabled || !element.visible {
		return
	}

	bounds := element.Bounds()
	if !bounds.Contains(oldPosition) {
		// Element was not hovered before so no need to send leave event or
		// process any children
		return
	}

	relativeNewPosition := newPosition.Translate(-bounds.X, -bounds.Y)
	relativeOldPosition := oldPosition.Translate(-bounds.X, -bounds.Y)
	if !bounds.Contains(newPosition) {
		element.onMouseEvent(MouseEvent{
			Position: relativeNewPosition,
			Type:     MouseEventTypeLeave,
		})
	}

	for childElement := element.lastChild; childElement != nil; childElement = childElement.leftSibling {
		w.processMouseLeave(childElement, relativeNewPosition, relativeOldPosition)
	}
}

func (w *windowHandler) processMouseEnter(element *Element, newPosition, oldPosition Position) {
	if !element.enabled || !element.visible {
		return
	}

	bounds := element.Bounds()
	if !bounds.Contains(newPosition) {
		// Element does not contain new point so no need to send enter event
		// or process any children
		return
	}

	relativeNewPosition := newPosition.Translate(-bounds.X, -bounds.Y)
	relativeOldPosition := oldPosition.Translate(-bounds.X, -bounds.Y)
	if !bounds.Contains(oldPosition) {
		element.onMouseEvent(MouseEvent{
			Position: relativeNewPosition,
			Type:     MouseEventTypeEnter,
		})
	}

	for childElement := element.lastChild; childElement != nil; childElement = childElement.leftSibling {
		w.processMouseEnter(childElement, relativeNewPosition, relativeOldPosition)
	}
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
			break // don't allow siblings that are underneath to process event
		}
	}

	// Let the current element handle the event.
	return element.onMouseEvent(event)
}

func (w *Window) renderElement(element *Element, canvas *Canvas, dirtyRegion Bounds) {
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
