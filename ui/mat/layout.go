package mat

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
)

// LayoutData represents a layout configuration for a component
// that is added to a Container.
type LayoutData struct {
	Left             optional.Int
	Right            optional.Int
	Top              optional.Int
	Bottom           optional.Int
	HorizontalCenter optional.Int
	VerticalCenter   optional.Int
	Width            optional.Int
	Height           optional.Int
	GrowHorizontally bool
	GrowVertically   bool
	Alignment        Alignment
}

// ElementLayoutData returns the LayoutData associated with
// the specified Element. If such has not been configured on the
// Element, then a default one is returned.
func ElementLayoutData(element *ui.Element) LayoutData {
	data, ok := element.LayoutConfig().(LayoutData)
	if !ok {
		return LayoutData{}
	}
	return data
}
