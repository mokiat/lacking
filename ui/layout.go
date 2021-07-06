package ui

import "strings"

var layoutRegistry map[string]LayoutBuilder

func init() {
	layoutRegistry = make(map[string]LayoutBuilder)
}

// Layout represents an algorithm through which Elements are
// positioned on the screen relative to their parents.
type Layout interface {

	// LayoutConfig creates a new layout config instance specific
	// to this layout.
	LayoutConfig() LayoutConfig

	// Apply applies this layout to the specified Element.
	Apply(element *Element)
}

// LayoutBuilder represents a mechanism through which
// Layouts can be created with specific properties.
type LayoutBuilder interface {

	// Build constructs a new Layout based on the specified
	// attributes.
	Build(attributes AttributeSet) (Layout, error)
}

// LayoutBuilderFunc is a helper function type that implements
// the LayoutBuilder interface.
type LayoutBuilderFunc func(attributes AttributeSet) (Layout, error)

// Build constructs a new Control instance.
func (f LayoutBuilderFunc) Build(attributes AttributeSet) (Layout, error) {
	return f(attributes)
}

// LayoutConfig represents a layout configuration for an Element.
// The actual implementation of this interface is determined by
// the parent Element's layout model.
type LayoutConfig interface {

	// ApplyAttributes configures this layout config
	// off of the specified attributes.
	ApplyAttributes(attributes AttributeSet)
}

// RegisterLayout adds the specified Layout implementation for
// the specified layout name.
func RegisterLayoutBuilder(name string, builder LayoutBuilder) {
	layoutRegistry[strings.ToLower(name)] = builder
}

// KnownLayoutName returns whether the specified layout
// name is know to this package.
//
// Use the RegisterLayout function to add new layout names.
func KnownLayoutName(name string) bool {
	_, ok := layoutRegistry[strings.ToLower(name)]
	return ok
}

// NamedLayout returns the Layout that is registered under the specified name.
func NamedLayoutBuilder(name string) (LayoutBuilder, bool) {
	builder, ok := layoutRegistry[strings.ToLower(name)]
	return builder, ok
}
