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
	l.applyTopToBottom(element)
	l.applyBottomToTop(element)
	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *VerticalLayout) applyTopToBottom(element *ui.Element) {
	contentBounds := element.ContentBounds()

	topPlacement := contentBounds.Y
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)
		if layoutConfig.Alignment == AlignmentBottom {
			continue
		}

		childBounds := ui.Bounds{
			Size: childElement.IdealSize(),
		}
		if layoutConfig.Width.Specified {
			childBounds.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childBounds.Height = layoutConfig.Height.Value
		}
		if layoutConfig.GrowHorizontally {
			childBounds.Width = contentBounds.Width
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

func (l *VerticalLayout) applyBottomToTop(element *ui.Element) {
	contentBounds := element.ContentBounds()

	bottomPlacement := contentBounds.Height
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)
		if layoutConfig.Alignment != AlignmentBottom {
			continue
		}

		childBounds := ui.Bounds{
			Size: childElement.IdealSize(),
		}
		if layoutConfig.Width.Specified {
			childBounds.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childBounds.Height = layoutConfig.Height.Value
		}
		if layoutConfig.GrowHorizontally {
			childBounds.Width = contentBounds.Width
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

		childBounds.Y = bottomPlacement - childBounds.Height
		childElement.SetBounds(childBounds)

		bottomPlacement -= childBounds.Height + l.contentSpacing
	}
}

func (l *VerticalLayout) calculateIdealSize(element *ui.Element) ui.Size {
	result := ui.NewSize(0, 0)
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}

		result.Width = maxInt(result.Width, childSize.Width)
		if result.Height > 0 {
			result.Height += l.contentSpacing
		}
		result.Height += childSize.Height
	}
	return result.Grow(element.Padding().Size())
}
