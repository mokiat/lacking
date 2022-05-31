package mat

import "github.com/mokiat/lacking/ui"

// AnchorLayoutSettings contains optional configurations for the
// AnchorLayout.
type AnchorLayoutSettings struct{}

// NewAnchorLayout creates a new AnchorLayout instance.
func NewAnchorLayout(settings AnchorLayoutSettings) *AnchorLayout {
	return &AnchorLayout{}
}

var _ ui.Layout = (*AnchorLayout)(nil)

// AnchorLayout is an implementation of Layout that positions and
// resizes elements according to border-relative coordinates.
type AnchorLayout struct{}

// Apply applies this layout to the specified Element.
func (l *AnchorLayout) Apply(element *ui.Element) {
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := childElement.LayoutConfig().(LayoutData)
		childBounds := ui.Bounds{
			Size: childElement.IdealSize(),
		}

		// horizontal
		if layoutConfig.Width.Specified {
			childBounds.Width = layoutConfig.Width.Value
			switch {
			case layoutConfig.Left.Specified:
				childBounds.X = l.leftPosition(element, layoutConfig.Left.Value)
			case layoutConfig.Right.Specified:
				childBounds.X = l.rightPosition(element, layoutConfig.Right.Value) - childBounds.Width
			case layoutConfig.HorizontalCenter.Specified:
				childBounds.X = l.horizontalCenterPosition(element, layoutConfig.HorizontalCenter.Value) - childBounds.Width/2
			}
		}

		// vertical
		if layoutConfig.Height.Specified {
			childBounds.Height = layoutConfig.Height.Value
			switch {
			case layoutConfig.Top.Specified:
				childBounds.Y = l.topPosition(element, layoutConfig.Top.Value)
			case layoutConfig.Bottom.Specified:
				childBounds.Y = l.bottomPosition(element, layoutConfig.Bottom.Value) - childBounds.Height
			case layoutConfig.VerticalCenter.Specified:
				childBounds.Y = l.verticalCenterPosition(element, layoutConfig.VerticalCenter.Value) - childBounds.Height/2
			}
		}

		if layoutConfig.Left.Specified {
			childBounds.X = l.leftPosition(element, layoutConfig.Left.Value)
			switch {
			case layoutConfig.Right.Specified:
				childBounds.Width = l.rightPosition(element, layoutConfig.Right.Value) - childBounds.X
			}
		}
		if layoutConfig.HorizontalCenter.Specified {
			childBounds.X = l.horizontalCenterPosition(element, layoutConfig.HorizontalCenter.Value) - childBounds.Width/2
		}

		if layoutConfig.Top.Specified {
			childBounds.Y = l.topPosition(element, layoutConfig.Top.Value)
			switch {
			case layoutConfig.Bottom.Specified:
				childBounds.Height = l.bottomPosition(element, layoutConfig.Bottom.Value) - childBounds.Y
			}
		}
		if layoutConfig.VerticalCenter.Specified {
			childBounds.Y = l.verticalCenterPosition(element, layoutConfig.VerticalCenter.Value) - childBounds.Height/2
		}

		childElement.SetBounds(childBounds)
	}

	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *AnchorLayout) calculateIdealSize(element *ui.Element) ui.Size {
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
		if layoutConfig.Left.Specified {
			childSize.Width += layoutConfig.Left.Value
		}
		if layoutConfig.Right.Specified {
			childSize.Width += layoutConfig.Right.Value
		}
		if layoutConfig.Top.Specified {
			childSize.Height += layoutConfig.Top.Value
		}
		if layoutConfig.Bottom.Specified {
			childSize.Height += layoutConfig.Bottom.Value
		}
		result.Width = maxInt(result.Width, childSize.Width)
		result.Height = maxInt(result.Height, childSize.Height)
	}
	return result.Grow(element.Padding().Size())
}

func (l *AnchorLayout) leftPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.X + value
}

func (l *AnchorLayout) rightPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.X + bounds.Width - value
}

func (l *AnchorLayout) topPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.Y + value
}

func (l *AnchorLayout) bottomPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.Y + bounds.Height - value
}

func (l *AnchorLayout) horizontalCenterPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.X + bounds.Width/2 + value
}

func (l *AnchorLayout) verticalCenterPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.Y + bounds.Height/2 + value
}
