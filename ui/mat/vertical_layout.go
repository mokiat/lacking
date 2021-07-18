package mat

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
)

// VerticalLayoutData represents a layout configuration for a component
// that is added to a Container with layout set to VerticalLayout.
type VerticalLayoutData struct {
	Width  optional.Int
	Height optional.Int
}

// VerticalLayoutSettings contains optional configurations for the
// VerticalLayout.
type VerticalLayoutSettings struct {
	ContentAlignment Alignment
	ContentSpacing   int
}

// NewVerticalLayout creates a new VerticalLayout instance.
func NewVerticalLayout(settings VerticalLayoutSettings) *VerticalLayout {
	return &VerticalLayout{
		contentAlignment: settings.ContentAlignment,
		contentSpacing:   settings.ContentSpacing,
	}
}

var _ ui.Layout = (*VerticalLayout)(nil)

// VerticalLayout is an implementation of Layout that positions and
// resizes elements down a vertical direction.
type VerticalLayout struct {
	contentAlignment Alignment
	contentSpacing   int
}

// Apply applies this layout to the specified Element.
func (l *VerticalLayout) Apply(element *ui.Element) {
	contentBounds := element.ContentBounds()

	topPlacement := contentBounds.Y
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := childElement.LayoutConfig().(VerticalLayoutData)

		childBounds := ui.Bounds{
			Size: childElement.IdealSize(),
		}
		if layoutConfig.Width.Specified {
			childBounds.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childBounds.Height = layoutConfig.Height.Value
		}

		switch l.contentAlignment {
		case AlignmentLeft:
			childBounds.X = contentBounds.X + childElement.Margin().Left
		case AlignmentRight:
			childBounds.X = contentBounds.X + contentBounds.Width - childElement.Margin().Right - childBounds.Width
		case AlignmentCenter:
			fallthrough
		default:
			childBounds.X = contentBounds.X + (contentBounds.Width-childBounds.Width)/2 - +childElement.Margin().Left
		}

		childBounds.Y = topPlacement + childElement.Margin().Top
		childElement.SetBounds(childBounds)

		topPlacement += childElement.Margin().Vertical() + childBounds.Height + l.contentSpacing
	}
}
