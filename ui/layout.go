package ui

// LayoutConfig represents a layout configuration for an Element.
// The actual implementation of this interface is determined by
// the parent Element's layout model.
type LayoutConfig interface {

	// ApplyAttributes configures this layout config
	// off of the specified attributes.
	ApplyAttributes(attributes AttributeSet)
}
