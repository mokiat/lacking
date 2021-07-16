package ui

import "reflect"

// Essence represents the behavior that is attached to an Element.
// For example, the actual value behind the interface could be a
// specific UI control and/or a handler.
type Essence interface{}

// ElementResizeHandler is a type of ElementHandler that can be
// used to receive events when an Element has been resized.
type ElementResizeHandler interface {
	OnResize(element *Element, bounds Bounds)
}

// ElementMouseHandler is a type of ElementHandler that can be
// used to receive events when the mouse has interacted with
// a given Element.
type ElementMouseHandler interface {
	OnMouseEvent(element *Element, event MouseEvent) bool
}

// ElementRenderHandler is a type of ElementHandler that can be
// used to receive events when a given Element is being
// rendered and to perform a custom rendering for the Element.
type ElementRenderHandler interface {
	OnRender(element *Element, canvas Canvas)
}

func newElement(context *Context) *Element {
	return &Element{
		context:      context,
		enabled:      true,
		visible:      true,
		materialized: true,
	}
}

// Element represents a hierarchical entity on the screen.
// It need not necessarily be a control and could just be
// an intermediate element used to group Controls.
type Element struct {
	id string

	parent       *Element
	firstChild   *Element
	lastChild    *Element
	rightSibling *Element
	leftSibling  *Element

	context *Context
	essence Essence

	enabled      bool
	visible      bool
	materialized bool

	margin       Spacing
	padding      Spacing
	bounds       Bounds
	idealSize    Size
	layoutConfig LayoutConfig
}

// ID returns the ID of this Element.
// If an ID was not specified, then an emptry string is returned.
func (e *Element) ID() string {
	return e.id
}

// SetID changes the ID of this Element. If this Element
// represents a Control, then this ID would affect the ID
// of that owner Control.
func (e *Element) SetID(id string) {
	e.id = id
}

// Parent returns the parent Element in the hierarchy. If this is
// the top-most Element then nil is returned.
func (e *Element) Parent() *Element {
	if e.parent == nil {
		return nil
	}
	return e.parent
}

// FirstChild returns the first (left-most) child Element of
// this Element. If this Element does not have any children
// then this method returns nil.
func (e *Element) FirstChild() *Element {
	if e.firstChild == nil {
		return nil
	}
	return e.firstChild
}

// LastChild returns the last (right-most) child Element of
// this Element. If this Element does not have any children
// then this method returns nil.
func (e *Element) LastChild() *Element {
	if e.lastChild == nil {
		return nil
	}
	return e.lastChild
}

// LeftSibling returns the left sibling Element of this Element.
// If this Element is the left-most child of its parent or does
// not have a parent then this method returns nil.
func (e *Element) LeftSibling() *Element {
	if e.leftSibling == nil {
		return nil
	}
	return e.leftSibling
}

// RightSibling returns the right sibling Element of this Element.
// If this Element is the right-most child of its parent or does
// not have a parent then this method returns nil.
func (e *Element) RightSibling() *Element {
	if e.rightSibling == nil {
		return nil
	}
	return e.rightSibling
}

// Detach removes this Element from the hierarchy but does not
// release any resources.
func (e *Element) Detach() {
	if e.parent != nil {
		e.parent.RemoveChild(e)
	}
}

// AppendSibling attaches an Element to the right of the current one.
func (e *Element) AppendSibling(sibling *Element) {
	sibling.Detach()
	sibling.rightSibling = e.rightSibling
	if e.rightSibling != nil {
		e.rightSibling.leftSibling = sibling
	}
	sibling.leftSibling = e
	e.rightSibling = sibling
	if e.parent != nil && sibling.rightSibling == nil {
		e.parent.lastChild = sibling
	}
}

// PrependChild adds the specified Element as the left-most child
// of this Element.
// If the preprended Element already has a parent, it is first
// detached from that parent.
func (e *Element) PrependChild(child *Element) {
	child.Detach()
	child.parent = e
	child.leftSibling = nil
	child.rightSibling = e.firstChild
	if e.firstChild != nil {
		e.firstChild.leftSibling = child
	}
	e.firstChild = child
	if e.lastChild == nil {
		e.lastChild = child
	}
}

// AppendChild adds the specified Element as the right-most child
// of this Element.
// If the appended Element already has a parent, it is first
// detached from that parent.
func (e *Element) AppendChild(child *Element) {
	child.Detach()
	child.parent = e
	child.leftSibling = e.lastChild
	child.rightSibling = nil
	if e.firstChild == nil {
		e.firstChild = child
	}
	if e.lastChild != nil {
		e.lastChild.rightSibling = child
	}
	e.lastChild = child
}

// RemoveChild removes the specified Element from the list of
// children held by this Element. If the specified Element is
// not a child of this Element, then nothing happens.
func (e *Element) RemoveChild(child *Element) {
	if child.parent != e {
		return
	}
	if child.leftSibling != nil {
		child.leftSibling.rightSibling = child.rightSibling
	}
	if child.rightSibling != nil {
		child.rightSibling.leftSibling = child.leftSibling
	}
	if e.firstChild == child {
		e.firstChild = child.rightSibling
	}
	if e.lastChild == child {
		e.lastChild = child.leftSibling
	}
	child.parent = nil
	child.leftSibling = nil
	child.rightSibling = nil
}

// Context returns the Context that is related to this Element's
// lifecycle.
func (e *Element) Context() *Context {
	return e.context
}

// Essence returns the Essence that is responsible for the behavior
// of this Element.
func (e *Element) Essence() Essence {
	return e.essence
}

// SetEssence changes the Essence that is to control the behavior
// and represent the purpose of this Element.
// Specifying nil indicates that this is a plain Element that will
// not handle events in any special way.
func (e *Element) SetEssence(essence Essence) {
	e.essence = essence
}

// As assigns the Essence of this Element to the target.
// If the target is not a pointer to the correct type, this
// method panics.
func (e *Element) As(target interface{}) {
	if target == nil {
		panic("target cannot be nil")
	}
	value := reflect.ValueOf(target)
	valueType := value.Type()
	if valueType.Kind() != reflect.Ptr {
		panic("target must be a pointer")
	}
	if value.IsNil() {
		panic("target pointer cannot be nil")
	}
	essenceType := reflect.TypeOf(e.essence)
	if !essenceType.AssignableTo(valueType.Elem()) {
		panic("cannot assign essence to specified type")
	}
	value.Elem().Set(reflect.ValueOf(e.essence))
}

// Margin returns the spacing that should be maintained
// around this Element. The margin does not reflect the
// Element's active area and is only a setting that is
// used by layouts to further adjust the Element's position.
func (e *Element) Margin() Spacing {
	return e.margin
}

// SetMargin sets the amount of space that should be left
// around the Element when positioned by a layout container.
func (e *Element) SetMargin(margin Spacing) {
	e.margin = margin
	// TODO: Trigger relayout
}

// Padding returns the spacing that should be maintained
// inside an Element between its outer bounds and its content
// area.
// This setting could affect the layouting of child Elements
// and clipping during rendering.
func (e *Element) Padding() Spacing {
	return e.padding
}

// SetPadding configures this Element's content area spacing.
func (e *Element) SetPadding(padding Spacing) {
	e.padding = padding
	// TODO: Trigger relayout
}

// Bounds returns the bounds of this Element relative
// to its parent. If it is a top-most Element, then
// the bounds are relative to the Window's content area.
func (e *Element) Bounds() Bounds {
	return e.bounds
}

// SetBounds configures this Element's bounds relative
// to its parent. If it is a top-most element, then
// the bounds are relative to the Window's content area.
func (e *Element) SetBounds(bounds Bounds) {
	e.bounds = bounds
	if resizeHandler, ok := e.essence.(ElementResizeHandler); ok {
		resizeHandler.OnResize(e, bounds)
	}
}

// ContentBounds returns the bounds of the content area
// of this Element relative to the Element itself.
// The content bounds are calculated based on the
// Element's bounds and padding.
func (e *Element) ContentBounds() Bounds {
	return Bounds{
		Position: NewPosition(
			e.padding.Left,
			e.padding.Top,
		),
		Size: NewSize(
			e.bounds.Width-e.padding.Horizontal(),
			e.bounds.Height-e.padding.Vertical(),
		),
	}
}

// IdealSize returns the ideal dimensions for this
// Element, which could (but not necessarily) be taken
// into consideration by layout containers.
func (e *Element) IdealSize() Size {
	return e.idealSize
}

// SetIdealSize changes this Element's ideal dimensions.
func (e *Element) SetIdealSize(size Size) {
	e.idealSize = size
	if e.parent != nil {
		if resizeHandler, ok := e.parent.essence.(ElementResizeHandler); ok {
			resizeHandler.OnResize(e.parent, e.parent.Bounds())
		}
	}
}

// LayoutConfig returns the layout configuration for
// this Element. The actual implementation of the
// LayoutConfig interface depends on the owner Element.
// If this Element does not have any layout preference,
// then nil is returned.
func (e *Element) LayoutConfig() LayoutConfig {
	return e.layoutConfig
}

// SetLayoutConfig changes this Element's layout configuration.
// The provided implementation should match the requirements
// of the parent layout Element, otherwise they will not be
// taken into consideration.
// If nil is specified, then default layouting should be used.
func (e *Element) SetLayoutConfig(layoutConfig LayoutConfig) {
	e.layoutConfig = layoutConfig
	if e.parent != nil {
		if resizeHandler, ok := e.parent.essence.(ElementResizeHandler); ok {
			resizeHandler.OnResize(e.parent, e.parent.Bounds())
		}
	}
}

// Enabled returns whether this element can be interacted with
// by the user.
func (e *Element) Enabled() bool {
	return e.enabled
}

// SetEnabled specifies whether this Element should be
// enabled for user interaction.
func (e *Element) SetEnabled(enabled bool) {
	e.enabled = enabled
	// TODO: Trigger redraw
}

// Visible returns whether this Element should be
// rendered and in respect whether it should receive events.
// Even if an Element is not Visible, it will still be
// considered by a layout. To fully remove an Element from
// the screen, the Materialized setting should be used.
func (e *Element) Visible() bool {
	return e.visible
}

// SetVisible controls whether this Element should be
// rendered.
func (e *Element) SetVisible(visible bool) {
	e.visible = visible
	// TODO: Trigger redraw
}

// Materialized controls whether this Element is at all
// present. If an Element is not materialized, then it is
// not rendered, does not receive events, and is not
// considered during layout evaluations. In essence, it is
// almost as though it has been removed, except that it
// is still part of the hierarchy.
func (e *Element) Materialized() bool {
	return e.materialized
}

// SetMaterialized specifies whether this Element should
// be considered in any way for rendering, events and
// layout calculations.
func (e *Element) SetMaterialized(materialized bool) {
	e.materialized = materialized
	if e.parent != nil {
		if resizeHandler, ok := e.parent.essence.(ElementResizeHandler); ok {
			resizeHandler.OnResize(e.parent, e.parent.Bounds())
		}
	}
	// TODO: Trigger redraw
}

// Destroy removes this Element from the hierarchy, as well
// as any child Elements and releases all allocated resources.
func (e *Element) Destroy() {
	e.Detach()
}

func (e *Element) onMouseEvent(event MouseEvent) bool {
	if mouseHandler, ok := e.essence.(ElementMouseHandler); ok {
		return mouseHandler.OnMouseEvent(e, event)
	}
	return false
}

func (e *Element) onRender(canvas Canvas) {
	if renderHandler, ok := e.essence.(ElementRenderHandler); ok {
		renderHandler.OnRender(e, canvas)
	}
}
