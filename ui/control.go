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

	// Element returns the hierarchy element that manages
	// this control. This method is intended for use by
	// internal logic and control implementations.
	// End users should not need to use this.
	Element() *Element
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
