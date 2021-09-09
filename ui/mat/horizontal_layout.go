package mat

import (
	"github.com/mokiat/lacking/ui"
)

// HorizontalLayoutSettings contains optional configurations for the
// HorizontalLayout.
type HorizontalLayoutSettings struct {
	ContentAlignment Alignment
	ContentSpacing   int
}

// NewHorizontalLayout creates a new HorizontalLayout instance.
func NewHorizontalLayout(settings HorizontalLayoutSettings) *HorizontalLayout {
	return &HorizontalLayout{
		contentAlignment: settings.ContentAlignment,
		contentSpacing:   settings.ContentSpacing,
	}
}

var _ ui.Layout = (*HorizontalLayout)(nil)

// HorizontalLayout is an implementation of Layout that positions and
// resizes elements in a horizontal direction.
type HorizontalLayout struct {
	contentAlignment Alignment
	contentSpacing   int
}

// Apply applies this layout to the specified Element.
func (l *HorizontalLayout) Apply(element *ui.Element) {
	contentBounds := element.ContentBounds()

	// fmt.Println()
	// fmt.Println("HORIZONTAL LAYOUT----")
	leftPlacement := contentBounds.X
	// fmt.Println("LEFT: ", leftPlacement)
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		// fmt.Println("CHILD!")
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
		case AlignmentTop:
			childBounds.Y = contentBounds.Y + childElement.Margin().Top
		case AlignmentBottom:
			childBounds.Y = contentBounds.Y + contentBounds.Height - childElement.Margin().Bottom - childBounds.Height
		case AlignmentCenter:
			fallthrough
		default:
			childBounds.Y = contentBounds.Y + (contentBounds.Height-childBounds.Height)/2 - childElement.Margin().Top
		}

		childBounds.X = leftPlacement + childElement.Margin().Left
		// fmt.Println("Bounds: ", childBounds)
		childElement.SetBounds(childBounds)

		leftPlacement += childElement.Margin().Horizontal() + childBounds.Width + l.contentSpacing
		// fmt.Println("LEFT: ", leftPlacement)
	}
}
