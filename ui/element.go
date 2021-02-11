package ui

// CreateElement creates a new Element.
func CreateElement() *Element {
	return &Element{
		visible: true,
		enabled: true,
	}
}

// ElementHandler is a mechanism by which an interested
// party (e.g. Control) can get notified of changes that
// occur to an Element.
type ElementHandler interface{}

// ResizeHandler is a mechanism by which an interested
// party (e.g. Control) can get notified of resize events that
// occur to an Element.
type ResizeHandler interface {
	OnResize(element *Element, bounds Bounds)
}

// RenderHandler is a mechanism by which an interested
// party (e.g. Control) can get notified of render events that
// occur to an Element.
type RenderHandler interface {
	OnRender(element *Element, ctx RenderContext)
}

// Element represents an hierarchical entity on the screen.
// It need not necessarily be a control and could just be
// an intermediate element used to group Controls.
type Element struct {
	parent       *Element
	firstChild   *Element
	lastChild    *Element
	rightSibling *Element
	leftSibling  *Element

	enabled      bool
	visible      bool
	materialized bool

	bounds  Bounds
	margin  Spacing
	padding Spacing

	control Control
	handler ElementHandler
}

// SetHandler changes the handler that will receive
// events regarding this element.
func (e *Element) SetHandler(handler ElementHandler) {
	e.handler = handler
}

// SetControl sets the control that is represented
// by this Element.
func (e *Element) SetControl(control Control) {
	e.control = control
}

// Control returns the Control that is represented by this
// Element. If this Element does not represent a Control,
// then this method returns nil.
func (e *Element) Control() Control {
	return e.control
}

// Parent returns the parent in the hierarchy. If this is
// the top-most Element then the method returns nil.
func (e *Element) Parent() *Element {
	return e.parent
}

// Append adds an Element as a child to this Element.
// The newly added Element will be placed last in the
// list of children.
func (e *Element) Append(element *Element) {
	if element.parent != nil {
		element.parent.Remove(element)
	}
	element.parent = e
	if e.firstChild == nil {
		e.firstChild = element
	}
	if e.lastChild != nil {
		e.lastChild.rightSibling = element
	}
	element.leftSibling = e.lastChild
	element.rightSibling = nil
	e.lastChild = element
}

// Remove removes the specified Element from the list
// of children of this Element.
func (e *Element) Remove(element *Element) {
	if element.parent != e {
		return
	}
	if element.leftSibling != nil {
		element.leftSibling.rightSibling = element.rightSibling
	}
	if element.rightSibling != nil {
		element.rightSibling.leftSibling = element.leftSibling
	}
	if element.parent.firstChild == element {
		element.parent.firstChild = element.rightSibling
	}
	if element.parent.lastChild == element {
		element.parent.lastChild = element.leftSibling
	}
	element.leftSibling = nil
	element.rightSibling = nil
	element.parent = nil
}

// FirstChild returns the first child Element of
// this Element.
func (e *Element) FirstChild() *Element {
	return e.firstChild
}

// LastChild returns the last child Element of
// this Element.
func (e *Element) LastChild() *Element {
	return e.lastChild
}

// LeftSibling returns the left sibling Element
// of this Element.
func (e *Element) LeftSibling() *Element {
	return e.leftSibling
}

// RightSibling returns the right sibling Element
// of this Element.
func (e *Element) RightSibling() *Element {
	return e.rightSibling
}

// SetBounds configures this Element's bounds relative
// to its parent. If it is a top-most element, then
// the bounds are relative to the window content area.
func (e *Element) SetBounds(bounds Bounds) {
	e.bounds = bounds
	if resizeHandler, ok := e.handler.(ResizeHandler); ok {
		resizeHandler.OnResize(e, bounds)
	}
}

// Bounds returns the bounds of this Element relative
// to its parent. If it is a top-most element, then
// the bounds are relative to the screen.
func (e *Element) Bounds() Bounds {
	return e.bounds
}

// ContentBounds returns the bounds of the content area
// of this Element relative to the Element's parent.
// The content area is determined by the Element's bounds
// and padding.
func (e *Element) ContentBounds() Bounds {
	return e.bounds.Translate(NewPosition(
		e.padding.Left,
		e.padding.Right,
	)).Shrink(NewSize(
		e.padding.Left+e.padding.Right,
		e.padding.Top+e.padding.Bottom,
	))
}

// SetMargin sets the amount of space that should be left
// around the element when placed in a layout
func (e *Element) SetMargin(margin Spacing) {
	e.margin = margin
	// TODO: Trigger relayout
}

// Margin returns the margin for this element.
func (e *Element) Margin() Spacing {
	return e.margin
}

// SetPadding sets the amount of space that should be left
// inside the component around the border when drawing child
// elements. This also defines the clipping box within which
// child Elements would be drawn.
func (e *Element) SetPadding(padding Spacing) {
	e.padding = padding
	// TODO: Trigger relayout
}

// Padding returns the padding for this element.
func (e *Element) Padding() Spacing {
	return e.padding
}

// SetEnabled controls whether this Element receives
// input events.
// Setting this to false means that the Element
// would not react to events like mouse or keyboard
// inputs and depending on the implementation might
// be rendered in a different way to indicate that
// it is not enabled.
func (e *Element) SetEnabled(enabled bool) {
	e.enabled = enabled
	// TODO: Trigger redraw
}

// IsEnabled returns whether the Element is to
// receive input events.
func (e *Element) IsEnabled() bool {
	return e.enabled
}

// SetVisible controls whether the Element should be
// displayed.
// Setting this to false does not cause Elements to
// be repositioned and instead it just prevents the
// Element from being rendered and from receiving
// input events.
func (e *Element) SetVisible(visible bool) {
	e.visible = visible
	// TODO: Trigger redraw
}

// IsVisible returns whether this Element should be
// rendered and receive events.
func (e *Element) IsVisible() bool {
	return e.visible
}

// SetMaterialized controls whether the Element is at
// all considered as existing.
// Setting this to false causes the Element to behave
// as though it has been deleted.
func (e *Element) SetMaterialized(materialized bool) {
	e.materialized = materialized
	// TODO: Trigger relayout
}

// IsMaterialized returns whether this Element is
// considered as existing or not.
func (e *Element) IsMaterialized() bool {
	return e.materialized
}
