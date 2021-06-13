package ui

import "strings"

var registry map[string]ControlBuilder

func init() {
	registry = make(map[string]ControlBuilder)
}

// Control represents an abstract user interface entity that
// the end user would interact with, in contrast to an Element,
// which is a building block of the hierarchy of Controls.
type Control interface {

	// ApplyAttributes applies the specified attributes
	// to this Control.
	ApplyAttributes(attributes AttributeSet) error

	// ID returns the unique identifier for the Control
	// if such has been specified, otherwise returns
	// an empty string.
	ID() string

	// SetID changes the ID of this Control and uderlying Element.
	//
	// The implementation may or may not check whether the ID is
	// already taken. Users should take care not to duplicate IDs
	// as otherwise the outcome is undefined.
	SetID(id string)

	// Element returns the hierarchy element that manages
	// this control. This method is intended for use by
	// internal logic and control implementations.
	// End users should not need to use this.
	Element() *Element

	// Context returns the lifecycle Context associated
	// with this Control.
	Context() *Context

	// Parent returns the parent control that holds this
	// control.
	// Note that the returned control is not necessary the
	// immediate parent element that references this Control.
	// It is possible for a parent Control to manage intermediate
	// Elements (e.g. rows, headers) that hold the control.
	// If this is the top-most Control, nil is returned.
	Parent() Control

	// Destroy removes this Control. It can no longer, and
	// should no longer, be used in any way.
	// If the Control is already deleted, this method
	// does nothing.
	Destroy()
}

// RegisterControlBuilder adds the specified ControlBuilder for
// the specified Control name.
func RegisterControlBuilder(name string, builder ControlBuilder) {
	registry[strings.ToLower(name)] = builder
}

// KnownControlName returns whether the specified Control
// name is know to this package.
//
// Use the Register function to add new Control names.
func KnownControlName(name string) bool {
	_, ok := registry[strings.ToLower(name)]
	return ok
}

// NamedControlBuilder returns the ControlBuilder that
// is registered under the specified name.
func NamedControlBuilder(name string) (ControlBuilder, bool) {
	builder, ok := registry[strings.ToLower(name)]
	return builder, ok
}

// ControlBuilder represents a mechanism through which
// Controls can be created from a Template.
type ControlBuilder interface {

	// Build constructs a new Control instance.
	Build(ctx *Context, template *Template, layoutConfig LayoutConfig) (Control, error)
}

// ControlBuilderFunc is a helper function type that implements
// the ControlBuilder interface.
type ControlBuilderFunc func(ctx *Context, template *Template, layoutConfig LayoutConfig) (Control, error)

// Build constructs a new Control instance.
func (f ControlBuilderFunc) Build(ctx *Context, template *Template, layoutConfig LayoutConfig) (Control, error) {
	return f(ctx, template, layoutConfig)
}

func newControl(element *Element) *baseControl {
	return &baseControl{
		element: element,
	}
}

var _ Control = (*baseControl)(nil)

type baseControl struct {
	element *Element
}

func (c *baseControl) ApplyAttributes(attributes AttributeSet) error {
	if stringValue, ok := attributes.StringAttribute("id"); ok {
		c.element.SetID(stringValue)
	}
	if boolValue, ok := attributes.BoolAttribute("enabled"); ok {
		c.element.SetEnabled(boolValue)
	}
	if boolValue, ok := attributes.BoolAttribute("visible"); ok {
		c.element.SetVisible(boolValue)
	}
	if boolValue, ok := attributes.BoolAttribute("materialized"); ok {
		c.element.SetMaterialized(boolValue)
	}

	margin := c.element.Margin()
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
	c.element.SetMargin(margin)

	padding := c.element.Padding()
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
	c.element.SetPadding(padding)
	return nil
}

func (c *baseControl) ID() string {
	return c.element.ID()
}

func (c *baseControl) SetID(id string) {
	c.element.SetID(id)
}

func (c *baseControl) Element() *Element {
	return c.element
}

func (c *baseControl) Context() *Context {
	return c.element.Context()
}

func (c *baseControl) Parent() Control {
	for pe := c.element.Parent(); pe != nil; pe = pe.Parent() {
		if control := pe.Control(); control != nil {
			return control
		}
	}
	return nil
}

func (c *baseControl) Destroy() {
	c.element.Destroy()
}
