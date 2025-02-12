package ui

import (
	"fmt"
	"time"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/app"
)

// newWindow creates a new Window instance that integrates
// with the specified app.Window.
func newWindow(appWindow app.Window, canvas *Canvas, resMan *resourceManager) (*Window, WindowHandler) {
	window := &Window{
		Window:             appWindow,
		canvas:             canvas,
		oldEnteredElements: make(map[*Element]struct{}),
		enteredElements:    make(map[*Element]struct{}),
		lastRender:         time.Now(),
	}
	window.context = newContext(nil, window, resMan)
	window.root = newElement(window)
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

	// TODO: Remove this interface and make it all internal

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

	// OnClipboardEvent is called whenever a clipboard paste operation
	// is being performed.
	OnClipboardEvent(event ClipboardEvent) bool

	// OnRender is called whenever the Window should redraw
	// itself.
	OnRender()

	// OnCloseRequested is called whenever the end-user has
	// indicated that they would like to close the application.
	// (e.g. using the close button on the application)
	OnCloseRequested() bool
}

// Window represents an application window.
type Window struct {
	app.Window
	context *Context
	canvas  *Canvas

	size           Size
	root           *Element
	focusedElement *Element

	oldEnteredElements map[*Element]struct{}
	enteredElements    map[*Element]struct{}

	oldMousePosition Position

	closeInterceptor func() bool

	lastRender time.Time
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

// CreateElement creates a new Element instance.
//
// The returned Element is not attached to anything and will not be
// drawn or processed in any way until it is attached to the Element
// hierarchy.
func (w *Window) CreateElement() *Element {
	return newElement(w)
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

// SetCloseInterceptor configures a handler to be notified of close
// requests by the user. This is not called on manual Close operations.
func (w *Window) SetCloseInterceptor(interceptor func() bool) {
	w.closeInterceptor = interceptor
}

// IsElementFocused returns whether the specified element is the currently
// focused Element.
func (w *Window) IsElementFocused(element *Element) bool {
	return w.focusedElement == element
}

// GrantFocus grants the focus to the specified Element.
func (w *Window) GrantFocus(element *Element) {
	w.focusedElement = element
	w.Invalidate()
}

// BubbleFocus releases focus from the current element and tries to find
// a parent that is focusable in order to grant it focus.
func (w *Window) BubbleFocus() {
	if w.focusedElement == nil {
		return
	}
	current := w.focusedElement.parent
	for current != nil {
		if current.focusable {
			w.focusedElement = current
			return
		}
		current = current.parent
	}
}

// DiscardFocus removes the focus from any Element.
func (w *Window) DiscardFocus() {
	w.focusedElement = nil
	w.Invalidate()
}

func (w *Window) Save() bool {
	currentElement := w.focusedElement
	for currentElement != nil {
		if currentElement.onSave() {
			return true
		}
		currentElement = currentElement.parent
	}
	return false
}

func (w *Window) Undo() bool {
	currentElement := w.focusedElement
	for currentElement != nil {
		if currentElement.onUndo() {
			return true
		}
		currentElement = currentElement.parent
	}
	return false
}

func (w *Window) Redo() bool {
	currentElement := w.focusedElement
	for currentElement != nil {
		if currentElement.onRedo() {
			return true
		}
		currentElement = currentElement.parent
	}
	return false
}

func (w *Window) Cut() bool {
	event := ClipboardEvent{
		Action: ClipboardActionCut,
	}
	currentElement := w.focusedElement
	for currentElement != nil {
		if currentElement.onClipboardEvent(event) {
			return true
		}
		currentElement = currentElement.parent
	}
	return false
}

func (w *Window) Copy() bool {
	event := ClipboardEvent{
		Action: ClipboardActionCopy,
	}
	currentElement := w.focusedElement
	for currentElement != nil {
		if currentElement.onClipboardEvent(event) {
			return true
		}
		currentElement = currentElement.parent
	}
	return false
}

func (w *Window) Paste() bool {
	w.RequestPaste()
	return true
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
		// TODO: The following check could be handled with mount and unmount
		// events and handling.
		if !current.HasAncestor(w.root) {
			w.DiscardFocus()
			return false
		}
		if current.focusable && current.onKeyboardEvent(event) {
			return true
		}
		current = current.parent
	}
	return false
}

func (w *windowHandler) OnMouseEvent(event MouseEvent) bool {
	w.checkMouseLeaveEnter(event.Position())
	w.oldMousePosition = event.Position()

	if event.Action == MouseActionDown {
		oldFocusedElement := w.focusedElement
		w.processFocusChange(w.root, event.Position())
		if w.focusedElement != oldFocusedElement {
			w.Invalidate()
		}
	}

	return w.processMouseEvent(w.root, event)
}

func (w *windowHandler) OnClipboardEvent(event ClipboardEvent) bool {
	current := w.focusedElement
	for current != nil {
		if current.focusable && current.onClipboardEvent(event) {
			return true
		}
		current = current.parent
	}
	return false
}

func (w *windowHandler) OnRender() {
	// Check that mouse is still in the same Element. This can change
	// when an Element gets disabled.
	w.checkMouseLeaveEnter(w.oldMousePosition)

	currentTime := time.Now()
	elapsedTime := currentTime.Sub(w.lastRender)
	if elapsedTime > time.Second {
		elapsedTime = time.Second
	}
	w.lastRender = currentTime

	if w.size.Empty() {
		return
	}

	clipBounds := Bounds{
		Position: NewPosition(0, 0),
		Size:     w.size,
	}
	dirtyRegion := clipBounds // TODO: Handle dirty sub-regions

	w.canvas.onBegin(elapsedTime)
	w.canvas.SetClipRect(
		0.0,
		float32(w.size.Width),
		0.0,
		float32(w.size.Height),
	)
	w.renderElement(w.root, w.canvas, clipBounds, dirtyRegion)
	w.canvas.onEnd()
}

func (w *windowHandler) OnCloseRequested() bool {
	if w.closeInterceptor != nil {
		return w.closeInterceptor()
	}
	return true
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

	relativePosition := position.Translate(bounds.Position.Inverse())
	for childElement := element.firstChild; childElement != nil; childElement = childElement.rightSibling {
		w.processFocusChange(childElement, relativePosition)
	}
}

func (w *windowHandler) checkMouseLeaveEnter(mousePosition Position) {
	w.oldEnteredElements, w.enteredElements = w.enteredElements, w.oldEnteredElements
	clear(w.enteredElements)

	w.processMouseLeave(w.root, mousePosition)
	w.processMouseLeaveInvisible(mousePosition)
	w.processMouseEnter(w.root, mousePosition)
}

func (w *windowHandler) processMouseLeave(element *Element, mousePosition Position) {
	if !element.visible {
		// We handle invisible elements separately.
		return
	}

	if _, ok := w.oldEnteredElements[element]; !ok {
		// Element was not hovered before so no need to send leave event or
		// process any children
		return
	}

	bounds := element.Bounds()
	relativeMousePosition := mousePosition.Translate(bounds.Position.Inverse())
	if !bounds.Contains(mousePosition) || !element.enabled {
		element.onMouseEvent(MouseEvent{
			Action: MouseActionLeave,
			X:      relativeMousePosition.X,
			Y:      relativeMousePosition.Y,
		})
	}

	for childElement := element.lastChild; childElement != nil; childElement = childElement.leftSibling {
		w.processMouseLeave(childElement, relativeMousePosition)
	}
}

func (w *windowHandler) processMouseLeaveInvisible(mousePosition Position) {
	for element := range w.oldEnteredElements {
		if !element.visible {
			bounds := element.AbsoluteBounds()
			relativeMousePosition := mousePosition.Translate(bounds.Position.Inverse())
			element.onMouseEvent(MouseEvent{
				Action: MouseActionLeave,
				X:      relativeMousePosition.X,
				Y:      relativeMousePosition.Y,
			})
		}
	}
}

func (w *windowHandler) processMouseEnter(element *Element, mousePosition Position) {
	if !element.enabled || !element.visible {
		return
	}

	bounds := element.Bounds()
	if !bounds.Contains(mousePosition) {
		// Element does not contain new point so no need to send enter event
		// or process any children
		return
	}

	relativeMousePosition := mousePosition.Translate(bounds.Position.Inverse())
	if _, ok := w.oldEnteredElements[element]; !ok {
		element.onMouseEvent(MouseEvent{
			Action: MouseActionEnter,
			X:      relativeMousePosition.X,
			Y:      relativeMousePosition.Y,
		})
	}
	w.enteredElements[element] = struct{}{}

	for childElement := element.lastChild; childElement != nil; childElement = childElement.leftSibling {
		w.processMouseEnter(childElement, relativeMousePosition)
	}
}

func (w *windowHandler) processMouseEvent(element *Element, event MouseEvent) bool {
	if !element.enabled || !element.visible {
		return false
	}

	// Check if any of the children (from top to bottom) can process the event.
	for childElement := element.lastChild; childElement != nil; childElement = childElement.leftSibling {
		if childBounds := childElement.Bounds(); childBounds.Contains(event.Position()) {
			translatedPosition := event.Position().Translate(childBounds.Position.Inverse())
			translatedEvent := event
			translatedEvent.X = translatedPosition.X
			translatedEvent.Y = translatedPosition.Y
			if w.processMouseEvent(childElement, translatedEvent) {
				return true
			}
			break // don't allow siblings that are underneath to process event
		}
	}

	// Let the current element handle the event.
	return element.onMouseEvent(event)
}

func (w *Window) renderElement(element *Element, canvas *Canvas, clipBounds, dirtyRegion Bounds) {
	if !element.visible {
		return
	}
	dirtyRegion = dirtyRegion.Intersect(element.bounds)
	if dirtyRegion.Empty() {
		return
	}
	dirtyRegion = dirtyRegion.Translate(element.bounds.Position.Inverse())

	elementClipBounds := clipBounds.
		Intersect(element.bounds).
		Translate(element.bounds.Position.Inverse())

	canvas.Push()
	canvas.Translate(sprec.NewVec2(
		float32(element.bounds.X),
		float32(element.bounds.Y),
	))
	canvas.SetClipRect(
		float32(elementClipBounds.X),
		float32(elementClipBounds.X+elementClipBounds.Width),
		float32(elementClipBounds.Y),
		float32(elementClipBounds.Y+elementClipBounds.Height),
	)
	element.onRender(canvas)
	if contentBounds := element.ContentBounds(); !contentBounds.Empty() {
		contentClipBounds := contentBounds.Intersect(elementClipBounds)
		canvas.SetClipRect(
			float32(contentClipBounds.X),
			float32(contentClipBounds.X+contentClipBounds.Width),
			float32(contentClipBounds.Y),
			float32(contentClipBounds.Y+contentClipBounds.Height),
		)
		for child := element.firstChild; child != nil; child = child.rightSibling {
			w.renderElement(
				child,
				canvas,
				contentClipBounds,
				dirtyRegion,
			)
		}
	}
	canvas.Pop()
}
