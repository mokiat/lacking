package layout

import "github.com/mokiat/lacking/ui"

// AnchorLayout is an implementation of Layout that positions and
// resizes elements according to border-relative coordinates.

// Anchor returns a layout that positions elements relative to key positions
// of the receiver element:
//   - left padding border
//   - right padding border
//   - top padding border
//   - bottom padding border
//   - horizontal center
//   - vertical center
//
// The exact positioning of a child is determined by its layout data. Individual
// children are positioned independently, hence two child elements could
// overlap.
func Anchor() ui.Layout {
	return &anchorLayout{}
}

type anchorLayout struct{}

func (l *anchorLayout) Apply(element *ui.Element) {
	contentBounds := element.ContentBounds()
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		idealSize := childElement.IdealSize()
		layoutConfig := ElementData(childElement)
		childBounds := ui.Bounds{}

		// Determine the width of the element.
		switch {
		case layoutConfig.GrowHorizontally:
			childBounds.Width = contentBounds.Width
		case layoutConfig.Width.Specified:
			childBounds.Width = layoutConfig.Width.Value
		case layoutConfig.Left.Specified && layoutConfig.Right.Specified:
			childBounds.Width = contentBounds.Width - (layoutConfig.Left.Value + layoutConfig.Right.Value)
		case layoutConfig.Left.Specified && layoutConfig.HorizontalCenter.Specified:
			childBounds.Width = (contentBounds.Width/2 + layoutConfig.HorizontalCenter.Value - layoutConfig.Left.Value) * 2
		case layoutConfig.Right.Specified && layoutConfig.HorizontalCenter.Specified:
			childBounds.Width = (contentBounds.Width/2 - layoutConfig.HorizontalCenter.Value - layoutConfig.Right.Value) * 2
		default:
			childBounds.Width = idealSize.Width
		}

		// Position the element horizontally.
		switch {
		case layoutConfig.Left.Specified:
			childBounds.X = l.leftPosition(contentBounds, layoutConfig.Left.Value)
		case layoutConfig.Right.Specified:
			childBounds.X = l.rightPosition(contentBounds, layoutConfig.Right.Value) - childBounds.Width
		case layoutConfig.HorizontalCenter.Specified:
			childBounds.X = l.horizontalCenterPosition(contentBounds, layoutConfig.HorizontalCenter.Value) - childBounds.Width/2
		default:
			childBounds.X = contentBounds.X // fallback to left-most
		}

		// Determine the height of the element.
		switch {
		case layoutConfig.GrowVertically:
			childBounds.Height = contentBounds.Height
		case layoutConfig.Height.Specified:
			childBounds.Height = layoutConfig.Height.Value
		case layoutConfig.Top.Specified && layoutConfig.Bottom.Specified:
			childBounds.Height = contentBounds.Height - (layoutConfig.Top.Value + layoutConfig.Bottom.Value)
		case layoutConfig.Top.Specified && layoutConfig.VerticalCenter.Specified:
			childBounds.Height = (contentBounds.Height/2 + layoutConfig.VerticalCenter.Value - layoutConfig.Top.Value) * 2
		case layoutConfig.Bottom.Specified && layoutConfig.VerticalCenter.Specified:
			childBounds.Height = (contentBounds.Height/2 - layoutConfig.VerticalCenter.Value - layoutConfig.Bottom.Value) * 2
		default:
			childBounds.Height = idealSize.Height
		}

		// Position the element vertically.
		switch {
		case layoutConfig.Top.Specified:
			childBounds.Y = l.topPosition(contentBounds, layoutConfig.Top.Value)
		case layoutConfig.Bottom.Specified:
			childBounds.Y = l.bottomPosition(contentBounds, layoutConfig.Bottom.Value) - childBounds.Height
		case layoutConfig.VerticalCenter.Specified:
			childBounds.Y = l.verticalCenterPosition(contentBounds, layoutConfig.VerticalCenter.Value) - childBounds.Height/2
		default:
			childBounds.Y = contentBounds.Y // fallback to top-most
		}

		childElement.SetBounds(childBounds)
	}

	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *anchorLayout) calculateIdealSize(element *ui.Element) ui.Size {
	result := ui.NewSize(0, 0)
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Left.Specified {
			childSize.Width += layoutConfig.Left.Value
		}
		if layoutConfig.Right.Specified {
			childSize.Width += layoutConfig.Right.Value
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
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

func (l *anchorLayout) leftPosition(contentBounds ui.Bounds, value int) int {
	return contentBounds.X + value
}

func (l *anchorLayout) rightPosition(contentBounds ui.Bounds, value int) int {
	return contentBounds.X + contentBounds.Width - value
}

func (l *anchorLayout) horizontalCenterPosition(contentBounds ui.Bounds, value int) int {
	return contentBounds.X + contentBounds.Width/2 + value
}

func (l *anchorLayout) topPosition(contentBounds ui.Bounds, value int) int {
	return contentBounds.Y + value
}

func (l *anchorLayout) bottomPosition(contentBounds ui.Bounds, value int) int {
	return contentBounds.Y + contentBounds.Height - value
}

func (l *anchorLayout) verticalCenterPosition(contentBounds ui.Bounds, value int) int {
	return contentBounds.Y + contentBounds.Height/2 + value
}
