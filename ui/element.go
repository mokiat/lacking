package ui

// Element represents a hierarchical entity on the screen.
// It need not necessarily be a control and could just be
// an intermediate element used to group Controls.
type ElementI interface {
	IdealSize() Size

	SetIdealSize(size Size)

	LayoutConfig() LayoutConfig
}

// ElementHandler is a mechanism by which an interested
// party (e.g. Control) can get notified of changes that
// occur to an Element.
type ElementHandler interface{}

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

type Element struct {
	id string

	parent       *Element
	firstChild   *Element
	lastChild    *Element
	rightSibling *Element
	leftSibling  *Element

	context *Context
	handler ElementHandler
	control Control

	enabled      bool
	visible      bool
	materialized bool

	margin       Spacing
	padding      Spacing
	bounds       Bounds
	idealSize    Size
	layoutConfig LayoutConfig
}

// ApplyAttributes applies the specified attributes
// to this Element.
func (e *Element) ApplyAttributes(attributes AttributeSet) error {
	if stringValue, ok := attributes.StringAttribute("id"); ok {
		e.SetID(stringValue)
	}
	if boolValue, ok := attributes.BoolAttribute("enabled"); ok {
		e.SetEnabled(boolValue)
	}
	if boolValue, ok := attributes.BoolAttribute("visible"); ok {
		e.SetVisible(boolValue)
	}
	if boolValue, ok := attributes.BoolAttribute("materialized"); ok {
		e.SetMaterialized(boolValue)
	}

	margin := e.Margin()
	if intValue, ok := attributes.IntAttribute("margin-top"); ok {
		margin.Top = intValue
	}
	if intValue, ok := attributes.IntAttribute("margin-bottom"); ok {
		margin.Bottom = intValue
	}
	if intValue, ok := attributes.IntAttribute("margin-left"); ok {
		margin.Left = intValue
	}
	if intValue, ok := attributes.IntAttribute("margin-right"); ok {
		margin.Right = intValue
	}
	e.SetMargin(margin)

	padding := e.Padding()
	if intValue, ok := attributes.IntAttribute("padding-top"); ok {
		padding.Top = intValue
	}
	if intValue, ok := attributes.IntAttribute("padding-bottom"); ok {
		padding.Bottom = intValue
	}
	if intValue, ok := attributes.IntAttribute("padding-left"); ok {
		padding.Left = intValue
	}
	if intValue, ok := attributes.IntAttribute("padding-right"); ok {
		padding.Right = intValue
	}
	e.SetPadding(padding)
	return nil
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

// PrependChild adds the specified Element as the left-most child
// of this Element.
// If the preprended Element already has a parent, it is first
// detached from that parent.
func (e *Element) PrependChild(child *Element) {
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}
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
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}
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

// Handler returns the ElementHandler that receives events for
// changes that occur to this Element or nil if none has been
// specified.
func (e *Element) Handler() ElementHandler {
	return e.handler
}

// SetHandler changes the ElementHandler that is to receive
// events regarding changes that occur to this Element.
func (e *Element) SetHandler(handler ElementHandler) {
	e.handler = handler
}

// Control returns the Control that is represented by this
// Element. If this Element does not represent a Control,
// then this method returns nil.
func (e *Element) Control() Control {
	return e.control
}

// SetControl sets the Control that is represented by this
// Element. If no Control should be associated with this
// Element, then a value of nil should be specified.
func (e *Element) SetControl(control Control) {
	e.control = control
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
	if resizeHandler, ok := e.handler.(ElementResizeHandler); ok {
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
		if resizeHandler, ok := e.parent.handler.(ElementResizeHandler); ok {
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
		if resizeHandler, ok := e.parent.handler.(ElementResizeHandler); ok {
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
		if resizeHandler, ok := e.parent.handler.(ElementResizeHandler); ok {
			resizeHandler.OnResize(e.parent, e.parent.Bounds())
		}
	}
	// TODO: Trigger redraw
}

// Destroy removes this Element from the hierarchy, as well
// as any child Elements and releases all allocated resources.
func (e *Element) Destroy() {
	if e.parent != nil {
		e.parent.RemoveChild(e)
	}
}

func (e *Element) onMouseEvent(event MouseEvent) bool {
	if mouseHandler, ok := e.handler.(ElementMouseHandler); ok {
		return mouseHandler.OnMouseEvent(e, event)
	}
	return false
}

func (e *Element) onRender(canvas Canvas) {
	if renderHandler, ok := e.handler.(ElementRenderHandler); ok {
		renderHandler.OnRender(e, canvas)
	}
}
