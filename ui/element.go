package ui

// Essence represents the behavior that is attached to an Element.
// For example, the actual value behind the interface could be a
// specific UI control and/or a handler.
type Essence interface{}

// ElementResizeHandler is a type of ElementHandler that can be
// used to receive events when an Element has been resized.
type ElementResizeHandler interface {
	OnResize(element *Element, bounds Bounds)
}

// ElementKeyboardHandler is a type of EventHandler that can be
// used to receive events when an element is focused and keyboard
// actions are performed.
type ElementKeyboardHandler interface {
	OnKeyboardEvent(element *Element, event KeyboardEvent) bool
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
	OnRender(element *Element, canvas *Canvas)
}

func newElement(window *Window) *Element {
	return &Element{
		window:    window,
		enabled:   true,
		visible:   true,
		focusable: false,
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
	leftSibling  *Element
	rightSibling *Element

	window  *Window
	essence Essence

	enabled   bool
	visible   bool
	focusable bool

	padding      Spacing
	bounds       Bounds
	idealSize    Size
	layout       Layout
	layoutConfig LayoutConfig
}

// ID returns the ID of this Element.
// If an ID was not specified, then an emptry string is returned.
func (e *Element) ID() string {
	return e.id
}

// SetID changes the ID of this Element.
func (e *Element) SetID(id string) {
	e.id = id
}

// Parent returns the parent Element in the hierarchy. If this is
// the top-most Element then nil is returned.
func (e *Element) Parent() *Element {
	return e.parent
}

// FirstChild returns the first (left-most) child Element of
// this Element. If this Element does not have any children
// then this method returns nil.
func (e *Element) FirstChild() *Element {
	return e.firstChild
}

// LastChild returns the last (right-most) child Element of
// this Element. If this Element does not have any children
// then this method returns nil.
func (e *Element) LastChild() *Element {
	return e.lastChild
}

// LeftSibling returns the left sibling Element of this Element.
// If this Element is the left-most child of its parent or does
// not have a parent then this method returns nil.
func (e *Element) LeftSibling() *Element {
	return e.leftSibling
}

// RightSibling returns the right sibling Element of this Element.
// If this Element is the right-most child of its parent or does
// not have a parent then this method returns nil.
func (e *Element) RightSibling() *Element {
	return e.rightSibling
}

// Detach removes this Element from the hierarchy but does not
// release any resources.
func (e *Element) Detach() {
	if e.parent != nil {
		e.parent.RemoveChild(e)
	}
}

// PrependSibling attaches an Element to the left of the current one.
func (e *Element) PrependSibling(sibling *Element) {
	sibling.Detach()
	sibling.leftSibling = e.leftSibling
	if e.leftSibling != nil {
		e.leftSibling.rightSibling = sibling
	}
	sibling.rightSibling = e
	e.leftSibling = sibling
	if e.parent != nil && sibling.leftSibling == nil {
		e.parent.firstChild = sibling
	}
	if e.parent != nil {
		e.parent.onBoundsChanged(e.parent.bounds)
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
	if e.parent != nil {
		e.parent.onBoundsChanged(e.parent.bounds)
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
	e.onBoundsChanged(e.bounds)
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
	e.onBoundsChanged(e.bounds)
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
	e.onBoundsChanged(e.bounds)
}

// Window returns the ui Window that owns this Element.
func (e *Element) Window() *Window {
	return e.window
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
	if padding != e.padding {
		e.padding = padding
		e.onBoundsChanged(e.bounds)
	}
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
	if bounds != e.bounds {
		e.bounds = bounds
		e.onBoundsChanged(bounds)
	}
}

// AbsoluteBounds returns the absolute bounds of the Element.
func (e *Element) AbsoluteBounds() Bounds {
	result := e.bounds
	for el := e.parent; el != nil; el = el.parent {
		result = result.Translate(el.bounds.Position)
	}
	return result
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
		Size: e.bounds.Size.Shrink(e.padding.Size()),
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
	if size != e.idealSize {
		e.idealSize = size
		if e.parent != nil {
			e.parent.onBoundsChanged(e.parent.bounds)
		}
	}
}

// Layout returns the layout configured for this Element.
func (e *Element) Layout() Layout {
	return e.layout
}

// SetLayout changes this Element's layout. If specified,
// this Element's children will be positioned according to the
// specified Layout.
// If nil is specified, then child Elements are not repositioned
// in any way and will use their configured bounds as positioning
// and size.
func (e *Element) SetLayout(layout Layout) {
	if layout != e.layout {
		e.layout = layout
		e.onBoundsChanged(e.bounds)
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
	if layoutConfig != e.layoutConfig {
		e.layoutConfig = layoutConfig
		if e.parent != nil {
			e.parent.onBoundsChanged(e.parent.bounds)
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
	if enabled != e.enabled {
		e.enabled = enabled
		e.Invalidate()
	}
}

// HierarchyEnabled checks whether this Element and all parent Elements
// are enabled.
func (e *Element) HierarchyEnabled() bool {
	if !e.enabled {
		return false
	}
	if e.parent == nil {
		return true
	}
	return e.parent.HierarchyEnabled()
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
	if visible != e.visible {
		e.visible = visible
		e.Invalidate()
	}
}

// HierarchyVisible checks whether this Element and all parent Elements
// are visible.
func (e *Element) HierarchyVisible() bool {
	if !e.visible {
		return false
	}
	if e.parent == nil {
		return true
	}
	return e.parent.HierarchyVisible()
}

// Focusable returns whether this Element can receive keyboard
// events. When an element if focusable and a mouse down event
// is received, it becomes the focused element and will begin
// to receive keyboard events.
func (e *Element) Focusable() bool {
	return e.focusable
}

// SetFocusable controls whether this Element should receive
// keyboard events.
func (e *Element) SetFocusable(focusable bool) {
	e.focusable = focusable
}

// Destroy removes this Element from the hierarchy, as well
// as any child Elements and releases all allocated resources.
func (e *Element) Destroy() {
	e.Detach()
}

// Invalidate marks this element as dirty and needing to be redrawn.
func (e *Element) Invalidate() {
	// TODO: Invalidate only the element's region
	e.window.Invalidate()
}

func (e *Element) onBoundsChanged(bounds Bounds) {
	if e.layout != nil {
		e.layout.Apply(e)
	}
	if resizeHandler, ok := e.essence.(ElementResizeHandler); ok {
		resizeHandler.OnResize(e, e.Bounds())
	}
	e.Invalidate()
}

func (e *Element) onKeyboardEvent(event KeyboardEvent) bool {
	if keyboardHandler, ok := e.essence.(ElementKeyboardHandler); ok {
		return keyboardHandler.OnKeyboardEvent(e, event)
	}
	return false
}

func (e *Element) onMouseEvent(event MouseEvent) bool {
	if mouseHandler, ok := e.essence.(ElementMouseHandler); ok {
		return mouseHandler.OnMouseEvent(e, event)
	}
	return false
}

func (e *Element) onRender(canvas *Canvas) {
	if renderHandler, ok := e.essence.(ElementRenderHandler); ok {
		renderHandler.OnRender(e, canvas)
	}
}
