package ui

// Control represents an abstract user interface
// control. Most methods on this interface should be
// safe for use from end users.
// Make sure you call the methods only from the UI
// thread as otherwise the behavior is unpredictable.
type Control interface {

	// ID returns the unique identifier for the Control
	// if such has been specified, otherwise returns
	// an empty string.
	ID() string

	// Element returns the hierarchy element that manages
	// this control. This method is intended for use by
	// internal logic and control implementations.
	// End users should not need to use this.
	Element() *Element

	// Parent returns the parent control that holds this
	// control.
	// Note that the returned control is not necessary the
	// immediate parent element that references this control.
	// It is possible for a parent control to manage intermediate
	// elements (e.g. rows, headers) that hold the control.
	// If this is the top-most control, nil is returned.
	Parent() Control

	// LayoutAttributes returns the layout attributes that
	// determine how this Control is positioned on screen.
	LayoutAttributes() AttributeSet

	// AddStyle adds a new style to this Control.
	// If the style has already been added, then this
	// operation does nothing.
	AddStyle(name string)

	// RemoveStyle removes a style from this Control.
	// If the style is not applied to the control, then
	// this method does nothing.
	RemoveStyle(name string)

	// HasStyle returns whether this Control has the
	// specified style assigned.
	HasStyle(name string) bool

	// Styles returns a slice of all styles applied to
	// this control.
	// Styles are a mechanism through which configurations
	// and layout data can be shared across controls.
	// They are applied in the order in which they were
	// added to the control, where latter styles override
	// former ones.
	Styles() []string

	// SetVisible controls whether the control should be
	// displayed.
	// Setting this to false does not cause controls to
	// be repositioned and instead it just prevents the
	// control from being rendered and from receiving
	// input events.
	SetVisible(visible bool)

	// IsVisible returns whether this control should be
	// rendered and receive events.
	IsVisible() bool

	// SetMaterialized controls whether the control is at
	// all considered as existing.
	// Setting this to false causes the control to behave
	// as though it has been deleted.
	SetMaterialized(materialized bool)

	// IsMaterialized returns whether this control is
	// considered as existing or not.
	IsMaterialized() bool

	// SetEnabled controls whether this control receives
	// input events.
	// Setting this to false means that the control
	// would not react to events like mouse or keyboard
	// inputs and depending on the implementation might
	// be rendered in a different way to indicate that
	// it is not enabled.
	SetEnabled(enabled bool)

	// IsEnabled returns whether the control is to
	// receive input events.
	IsEnabled() bool

	// Delete removes this control. It can no longer and
	// should no longer be used in any way.
	// If the control is already deleted, this method
	// does nothing.
	Delete()
}

func CreateControl(id string, element *Element, layoutAttributes AttributeSet) Control {
	return &control{
		id:               id,
		element:          element,
		layoutAttributes: layoutAttributes,
	}
}

type control struct {
	id               string
	element          *Element
	styles           []string     // come from template, not attrubutes
	layoutAttributes AttributeSet // come from template but have a mutable layer on top
}

func (c *control) ID() string {
	return c.id
}

func (c *control) Element() *Element {
	return c.element
}

func (c *control) Parent() Control {
	parentElement := c.element.Parent()
	for parentElement != nil {
		if parentControl := parentElement.Control(); parentControl != nil {
			return parentControl
		}
		parentElement = parentElement.Parent()
	}
	return nil
}

func (c *control) LayoutAttributes() AttributeSet {
	return c.layoutAttributes
}

func (c *control) AddStyle(name string) {
	styleIndex := c.styleIndex(name)
	if styleIndex != -1 {
		return
	}
	c.styles = append(c.styles, name)
}

func (c *control) RemoveStyle(name string) {
	styleIndex := c.styleIndex(name)
	if styleIndex == -1 {
		return
	}
	for i := styleIndex; i < len(c.styles)-1; i++ {
		c.styles[i] = c.styles[i+1]
	}
	c.styles = c.styles[:len(c.styles)-1]
}

func (c *control) HasStyle(name string) bool {
	return c.styleIndex(name) != -1
}

func (c *control) Styles() []string {
	stylesCopy := make([]string, len(c.styles))
	copy(stylesCopy, c.styles)
	return stylesCopy
}

func (c *control) SetVisible(visible bool) {
	c.element.SetVisible(visible)
}

func (c *control) IsVisible() bool {
	return c.element.IsVisible()
}

func (c *control) SetMaterialized(materialized bool) {
	c.element.SetMaterialized(materialized)
}

func (c *control) IsMaterialized() bool {
	return c.element.IsMaterialized()
}

func (c *control) SetEnabled(enabled bool) {
	c.element.SetEnabled(enabled)
}

func (c *control) IsEnabled() bool {
	return c.element.IsEnabled()
}

func (c *control) Delete() {
	if c.element.Parent() != nil {
		c.element.Parent().Remove(c.element)
	}
}

func (c *control) styleIndex(name string) int {
	for i, style := range c.styles {
		if style == name {
			return i
		}
	}
	return -1
}
