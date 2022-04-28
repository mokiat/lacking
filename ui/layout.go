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

// NewFillLayout returns a new FillLayout instance.
func NewFillLayout() *FillLayout {
	return &FillLayout{}
}

var _ Layout = (*FillLayout)(nil)

// FillLayout resizes the children to fill the content space
// of the parent element.
type FillLayout struct{}

// Apply applies this layout to the specified Element.
func (l *FillLayout) Apply(element *Element) {
	for child := element.FirstChild(); child != nil; child = child.RightSibling() {
		child.SetBounds(element.ContentBounds())
	}
}
