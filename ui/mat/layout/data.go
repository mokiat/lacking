package layout

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
)

// Data represents a layout configuration for a component that is added to a
// layout container.
type Data struct {

	// Left indicates the positioning relative to the parent's left border.
	Left opt.T[int]

	// Right indicates the positioning relative to the parent's right border.
	Right opt.T[int]

	// Top indicates the positioning relative to the parent's top border.
	Top opt.T[int]

	// Bottom indicates the positioning relative to the parent's bottom border.
	Bottom opt.T[int]

	// HorizontalCenter indicates the positioning relative to the parent's
	// horizontal center.
	HorizontalCenter opt.T[int]

	// VerticalCenter indicates the positioning relative to the parent's
	// vertical center.
	VerticalCenter opt.T[int]

	// Width specifies the desired width of the component.
	Width opt.T[int]

	// Height specifies the desired height of the component.
	Height opt.T[int]

	// GrowHorizontally indicates whether the component should occupy as much
	// space horizontally as possible.
	GrowHorizontally bool

	// GrowVertically indicates whether the component should occupy as much
	// space vertically as possible.
	GrowVertically bool

	// HorizontalAlignment indicates how a component should be aligned
	// horizontally if this cannot be determined by any other parameters.
	HorizontalAlignment HorizontalAlignment

	// VerticalAlignment indicates how a component should be aligned
	// vertically if this cannot be determined by any other parameters.
	VerticalAlignment VerticalAlignment
}

// ElementData returns the layout data associated with the specified element.
// If one has not been assigned to the element or if it is of a different
// type then a default one is returned.
func ElementData(element *ui.Element) Data {
	data, ok := element.LayoutConfig().(Data)
	if !ok {
		return Data{}
	}
	return data
}
