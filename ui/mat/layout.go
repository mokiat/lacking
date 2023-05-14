package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
)

// LayoutData represents a layout configuration for a component
// that is added to a container.
type LayoutData struct {
	Left             opt.T[int]
	Right            opt.T[int]
	Top              opt.T[int]
	Bottom           opt.T[int]
	HorizontalCenter opt.T[int]
	VerticalCenter   opt.T[int]
	Width            opt.T[int]
	Height           opt.T[int]
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
