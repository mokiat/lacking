package ui

// Layout represents an algorithm through which child Elements are
// positioned on the screen relative to their parent.
type Layout interface {

	// Apply applies this layout to the specified Element.
	Apply(element *Element)
}

// LayoutConfig represents a layout configuration for an Element.
// The actual implementation of this interface is determined by
// the parent Element's layout model.
type LayoutConfig interface{}
