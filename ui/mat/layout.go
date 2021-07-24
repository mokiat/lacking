package mat

import "github.com/mokiat/lacking/ui/optional"

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
}
