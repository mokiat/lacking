package ui

import "fmt"

const (
	// ViewModeNone indicates no preferences.
	ViewModeNone ViewMode = iota

	// ViewModeReplace indicates that the existing
	// View should be replaced. When the new View is
	// closed, the old View would not exist to take
	// back its place.
	ViewModeReplace

	// ViewModeCover indicates that the new View
	// should cover the old View, fully occluding it.
	// Once the new View is closed, the former View
	// is shown back to the screen.
	ViewModeCover

	// ViewModeOverlay indicates that the new View should
	// cover the old View but the old View should continue
	// to be drawn as the new View may just partially
	// cover the old one.
	ViewModeOverlay
)

// ViewMode indicates how a View should be opened
// and what happens when another View is opened
// afterwards.
type ViewMode int

// Is checks whether the given ViewMode contains
// the specified ViewMode flag.
func (m ViewMode) Is(flag ViewMode) bool {
	return (m & flag) != 0
}

// ViewType represents a factory that can initialize
// Views in a particular fashion.
type ViewType interface {

	// SetupView initializes the specified View.
	SetupView(view *View) error
}

// ViewHandler is a mechanism by which the framework can
// communicate key events to the user-defined View.
type ViewHandler interface {

	// OnCreate is called whenever a View should prepare itself
	// for visualization.
	OnCreate(view *View)

	// OnShow is called prior to the View being shown on the
	// screen.
	OnShow(view *View)

	// OnHide is called prior to the View being hidden from
	// the screen.
	OnHide(view *View)

	// OnDestroy is called whenever a View is about to be
	// destroyed and should release anything it has allocated.
	OnDestroy(view *View)
}

func newView(window *Window, context *Context) *View {
	return &View{
		window:  window,
		context: context,
	}
}

// View represents an entity on the screen that usually
// occupies the whole screen and has a dedicated lifecycle.
type View struct {
	window         *Window
	context        *Context
	root           Control
	handler        ViewHandler
	pointedElement *Element
}

// Window returns the Window that holds this View.
func (v *View) Window() *Window {
	return v.window
}

// Context returns the Context for this View. Elements
// within this View will normally have a Context that
// inherits from this one.
func (v *View) Context() *Context {
	return v.context
}

// Root returns the top-most Control of this View.
func (v *View) Root() Control {
	return v.root
}

// SetRoot changes the top-most Control of this View.
func (v *View) SetRoot(root Control) {
	v.root = root
}

// Handler returns the ViewHandler that manages this View.
func (v *View) Handler() ViewHandler {
	return v.handler
}

// SetHandler changes the ViewHandler that manages this View.
func (v *View) SetHandler(handler ViewHandler) {
	v.handler = handler
}

// FindElementByID looks up the Element hierarchy tree for an Element
// with the specified ID.
func (v *View) FindElementByID(id string) (*Element, bool) {
	if v.root == nil {
		return nil, false
	}
	rootElement := v.root.Element()

	return v.dfsElementByID(rootElement, id)
}

// GetElementByID looks up the Element hierarchy tree for an Element
// with the specified ID.
func (v *View) GetElementByID(id string) *Element {
	element, found := v.FindElementByID(id)
	if !found {
		panic(fmt.Errorf("element with id %q not found", id))
	}
	return element
}

// Close closes this view and releases all resources
// allocated to it.
func (v *View) Close() {
	panic("TODO")
}

func (v *View) onCreate() {
	if v.handler != nil {
		v.handler.OnCreate(v)
	}
}

func (v *View) onShow() {
	if v.handler != nil {
		v.handler.OnShow(v)
	}
}

func (v *View) onHide() {
	if v.handler != nil {
		v.handler.OnHide(v)
	}
}

func (v *View) onDestroy() {
	if v.handler != nil {
		v.handler.OnDestroy(v)
	}
}

func (v *View) onResize(size Size) {
	if v.root != nil {
		v.root.Element().SetBounds(Bounds{
			Position: NewPosition(0, 0),
			Size:     size,
		})
	}
}

func (v *View) onKeyboardEvent(event KeyboardEvent) bool {
	return false
}

func (v *View) onMouseEvent(event MouseEvent) bool {
	if v.root != nil {
		rootElement := v.root.Element()
		return v.processMouseEvent(rootElement, event)
	}
	return false
}

func (v *View) onRender(canvas Canvas, dirtyRegion Bounds) {
	if v.root != nil {
		rootElement := v.root.Element()
		v.renderElement(rootElement, canvas, dirtyRegion)
	}
}

func (v *View) dfsElementByID(current *Element, id string) (*Element, bool) {
	if current == nil {
		return nil, false
	}
	if current.id == id {
		return current, true
	}
	for child := current.firstChild; child != nil; child = child.rightSibling {
		if result, found := v.dfsElementByID(child, id); found {
			return result, true
		}
	}
	return nil, false
}

func (v *View) processMouseEvent(element *Element, event MouseEvent) bool {
	// Check if any of the children (from top to bottom) can process the event.
	for childElement := element.lastChild; childElement != nil; childElement = childElement.leftSibling {
		if childBounds := childElement.Bounds(); childBounds.Contains(event.Position) {
			event.Position = event.Position.Translate(-childBounds.X, -childBounds.Y)
			return v.processMouseEvent(childElement, event)
		}
	}

	// Check if we need to change mouse ownership.
	if element != v.pointedElement {
		if v.pointedElement != nil {
			v.pointedElement.onMouseEvent(MouseEvent{
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

		v.pointedElement = element
	}

	// Let the current element handle the event.
	return element.onMouseEvent(event)
}

func (v *View) renderElement(element *Element, canvas Canvas, dirtyRegion Bounds) {
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
			v.renderElement(child, canvas, dirtyRegion.Translate(element.bounds.Position.Inverse()))
		}
	}
	canvas.Pop()
}
