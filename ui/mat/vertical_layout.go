package mat

import "github.com/mokiat/lacking/ui"

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
		layoutConfig := childElement.LayoutConfig().(LayoutData)

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
			childBounds.X = contentBounds.X
		case AlignmentRight:
			childBounds.X = contentBounds.X + contentBounds.Width - childBounds.Width
		case AlignmentCenter:
			fallthrough
		default:
			childBounds.X = contentBounds.X + (contentBounds.Width-childBounds.Width)/2
		}

		childBounds.Y = topPlacement
		childElement.SetBounds(childBounds)

		topPlacement += childBounds.Height + l.contentSpacing
	}
}
