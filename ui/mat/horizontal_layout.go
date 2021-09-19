package mat

import (
	"github.com/mokiat/lacking/ui"
)

// HorizontalLayoutSettings contains optional configurations for the
// HorizontalLayout.
type HorizontalLayoutSettings struct {
	ContentAlignment Alignment
	ContentSpacing   int
	Flipped          bool
}

// NewHorizontalLayout creates a new HorizontalLayout instance.
func NewHorizontalLayout(settings HorizontalLayoutSettings) *HorizontalLayout {
	return &HorizontalLayout{
		contentAlignment: settings.ContentAlignment,
		contentSpacing:   settings.ContentSpacing,
		flipped:          settings.Flipped,
	}
}

var _ ui.Layout = (*HorizontalLayout)(nil)

// HorizontalLayout is an implementation of Layout that positions and
// resizes elements in a horizontal direction.
type HorizontalLayout struct {
	contentAlignment Alignment
	contentSpacing   int
	flipped          bool
}

// Apply applies this layout to the specified Element.
func (l *HorizontalLayout) Apply(element *ui.Element) {
	if l.flipped {
		l.applyRightToLeft(element)
	} else {
		l.applyLeftToRight(element)
	}
	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *HorizontalLayout) applyLeftToRight(element *ui.Element) {
	contentBounds := element.ContentBounds()

	leftPlacement := contentBounds.X
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

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
		case AlignmentTop:
			childBounds.Y = contentBounds.Y
		case AlignmentBottom:
			childBounds.Y = contentBounds.Y + contentBounds.Height - childBounds.Height
		case AlignmentCenter:
			fallthrough
		default:
			childBounds.Y = contentBounds.Y + (contentBounds.Height-childBounds.Height)/2
		}

		childBounds.X = leftPlacement
		childElement.SetBounds(childBounds)

		leftPlacement += childBounds.Width + l.contentSpacing
	}
}

func (l *HorizontalLayout) applyRightToLeft(element *ui.Element) {
	contentBounds := element.ContentBounds()

	rightPlacement := contentBounds.Width
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

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
		case AlignmentTop:
			childBounds.Y = contentBounds.Y
		case AlignmentBottom:
			childBounds.Y = contentBounds.Y + contentBounds.Height - childBounds.Height
		case AlignmentCenter:
			fallthrough
		default:
			childBounds.Y = contentBounds.Y + (contentBounds.Height-childBounds.Height)/2
		}

		childBounds.X = rightPlacement - childBounds.Width
		childElement.SetBounds(childBounds)

		rightPlacement -= childBounds.Width + l.contentSpacing
	}
}

func (l *HorizontalLayout) calculateIdealSize(element *ui.Element) ui.Size {
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

		result.Height = maxInt(result.Height, childSize.Height)
		if result.Width > 0 {
			result.Width += l.contentSpacing
		}
		result.Width += childSize.Width
	}
	result.Width += element.Padding().Horizontal()
	result.Height += element.Padding().Vertical()
	return result
}
