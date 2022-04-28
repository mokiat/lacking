package mat

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/util/optional"
)

// LayoutData represents a layout configuration for a component
// that is added to a Container.
type LayoutData struct {
	Left             optional.V[int]
	Right            optional.V[int]
	Top              optional.V[int]
	Bottom           optional.V[int]
	HorizontalCenter optional.V[int]
	VerticalCenter   optional.V[int]
	Width            optional.V[int]
	Height           optional.V[int]
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
