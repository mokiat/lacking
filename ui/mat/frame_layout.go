package mat

import "github.com/mokiat/lacking/ui"

// NewFrameLayout creates a new FrameLayout instance.
func NewFrameLayout() *FrameLayout {
	return &FrameLayout{}
}

var _ ui.Layout = (*FrameLayout)(nil)

// FrameLayout is an implementation of Layout that positions and
// resizes elements in a frame form fashion.
type FrameLayout struct{}

// Apply applies this layout to the specified Element.
func (l *FrameLayout) Apply(element *ui.Element) {
	leftSize := ui.Size{}
	rightSize := ui.Size{}
	topSize := ui.Size{}
	bottomSize := ui.Size{}
	centerSize := ui.Size{}

	// During first iteration we calculate ideal border sizes
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}

		switch layoutConfig.Alignment {
		case AlignmentCenter:
			centerSize = ui.Size{
				Width:  maxInt(centerSize.Width, childSize.Width),
				Height: maxInt(centerSize.Height, childSize.Height),
			}
		case AlignmentLeft:
			leftSize = ui.Size{
				Width:  maxInt(leftSize.Width, childSize.Width),
				Height: maxInt(leftSize.Height, childSize.Height),
			}
		case AlignmentRight:
			rightSize = ui.Size{
				Width:  maxInt(rightSize.Width, childSize.Width),
				Height: maxInt(rightSize.Height, childSize.Height),
			}
		case AlignmentTop:
			topSize = ui.Size{
				Width:  maxInt(topSize.Width, childSize.Width),
				Height: maxInt(topSize.Height, childSize.Height),
			}
		case AlignmentBottom:
			bottomSize = ui.Size{
				Width:  maxInt(bottomSize.Width, childSize.Width),
				Height: maxInt(bottomSize.Height, childSize.Height),
			}
		}
	}

	// Store ideal size but don't set yet
	idealSize := ui.Size{
		Width: maxInt(
			maxInt(topSize.Width, bottomSize.Width),
			leftSize.Width+centerSize.Width+rightSize.Width,
		) + element.Padding().Horizontal(),
		Height: topSize.Height + bottomSize.Height + maxInt(
			maxInt(leftSize.Height, rightSize.Height),
			centerSize.Height,
		) + element.Padding().Vertical(),
	}

	contentBounds := element.ContentBounds()

	// We don't allow borders to extend more than half the content area
	topSize = ui.Size{
		Width:  contentBounds.Width,
		Height: minInt(topSize.Height, contentBounds.Height/2),
	}
	bottomSize = ui.Size{
		Width:  contentBounds.Width,
		Height: minInt(bottomSize.Height, contentBounds.Height/2),
	}
	leftSize = ui.Size{
		Width:  minInt(leftSize.Width, contentBounds.Width/2),
		Height: contentBounds.Height - topSize.Height - bottomSize.Height,
	}
	rightSize = ui.Size{
		Width:  minInt(rightSize.Width, contentBounds.Width/2),
		Height: contentBounds.Height - topSize.Height - bottomSize.Height,
	}
	centerSize = ui.Size{
		Width:  contentBounds.Width - leftSize.Width - rightSize.Width,
		Height: contentBounds.Height - topSize.Height - bottomSize.Height,
	}

	// During second iteration we actually layout the children
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

		switch layoutConfig.Alignment {
		case AlignmentCenter:
			childElement.SetBounds(ui.Bounds{
				Position: ui.NewPosition(
					contentBounds.X+leftSize.Width,
					contentBounds.Y+topSize.Height,
				),
				Size: centerSize,
			})
		case AlignmentLeft:
			childElement.SetBounds(ui.Bounds{
				Position: ui.NewPosition(
					contentBounds.X,
					contentBounds.Y+topSize.Height,
				),
				Size: leftSize,
			})
		case AlignmentRight:
			childElement.SetBounds(ui.Bounds{
				Position: ui.NewPosition(
					contentBounds.X+leftSize.Width+centerSize.Width,
					contentBounds.Y+topSize.Height,
				),
				Size: rightSize,
			})
		case AlignmentTop:
			childElement.SetBounds(ui.Bounds{
				Position: ui.NewPosition(
					contentBounds.X,
					contentBounds.Y,
				),
				Size: topSize,
			})
		case AlignmentBottom:
			childElement.SetBounds(ui.Bounds{
				Position: ui.NewPosition(
					contentBounds.X,
					contentBounds.Y+topSize.Height+centerSize.Height,
				),
				Size: bottomSize,
			})
		}
	}

	// Store ideal size, which may retrigger the layout
	element.SetIdealSize(idealSize)
}
