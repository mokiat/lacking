package layout

import "github.com/mokiat/lacking/ui"

// Frame returns a layout that positions elements around a frame (like a picture
// frame). The five main sections are top,left,center,right,bottom and are
// distributed as follows.
//
//	 _____________
//	|______T______|
//	|   |     |   |
//	| L |  C  | R |
//	|___|_____|___|
//	|______B______|
func Frame() ui.Layout {
	return &frameLayout{}
}

type frameLayout struct{}

func (l *frameLayout) Apply(element *ui.Element) {
	var (
		topHeight    int
		bottomHeight int
		leftWidth    int
		rightWidth   int
	)

	// During first iteration we calculate border sizes.
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}

		switch layoutConfig.VerticalAlignment {
		case VerticalAlignmentTop:
			topHeight = maxInt(topHeight, childSize.Height)
		case VerticalAlignmentBottom:
			bottomHeight = maxInt(bottomHeight, childSize.Height)
		default: // treat as center
			switch layoutConfig.HorizontalAlignment {
			case HorizontalAlignmentLeft:
				leftWidth = maxInt(leftWidth, childSize.Width)
			case HorizontalAlignmentRight:
				rightWidth = maxInt(rightWidth, childSize.Width)
			}
		}
	}

	// NOTE: We don't allow borders to extend more than half the content area.
	contentBounds := element.ContentBounds()
	topSize := ui.Size{
		Width:  contentBounds.Width,
		Height: minInt(topHeight, contentBounds.Height/2),
	}
	bottomSize := ui.Size{
		Width:  contentBounds.Width,
		Height: minInt(bottomHeight, contentBounds.Height/2),
	}
	leftSize := ui.Size{
		Width:  minInt(leftWidth, contentBounds.Width/2),
		Height: contentBounds.Height - topSize.Height - bottomSize.Height,
	}
	rightSize := ui.Size{
		Width:  minInt(rightWidth, contentBounds.Width/2),
		Height: contentBounds.Height - topSize.Height - bottomSize.Height,
	}
	centerSize := ui.Size{
		Width:  contentBounds.Width - leftSize.Width - rightSize.Width,
		Height: contentBounds.Height - topSize.Height - bottomSize.Height,
	}

	// During second iteration we actually layout the children.
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)

		switch layoutConfig.VerticalAlignment {
		case VerticalAlignmentTop:
			childElement.SetBounds(ui.Bounds{
				Position: ui.NewPosition(
					contentBounds.X,
					contentBounds.Y,
				),
				Size: topSize,
			})
		case VerticalAlignmentBottom:
			childElement.SetBounds(ui.Bounds{
				Position: ui.NewPosition(
					contentBounds.X,
					contentBounds.Y+topSize.Height+centerSize.Height,
				),
				Size: bottomSize,
			})
		default: // treat as center
			switch layoutConfig.HorizontalAlignment {
			case HorizontalAlignmentLeft:
				childElement.SetBounds(ui.Bounds{
					Position: ui.NewPosition(
						contentBounds.X,
						contentBounds.Y+topSize.Height,
					),
					Size: leftSize,
				})
			case HorizontalAlignmentRight:
				childElement.SetBounds(ui.Bounds{
					Position: ui.NewPosition(
						contentBounds.X+leftSize.Width+centerSize.Width,
						contentBounds.Y+topSize.Height,
					),
					Size: rightSize,
				})
			default: // treat as center
				childElement.SetBounds(ui.Bounds{
					Position: ui.NewPosition(
						contentBounds.X+leftSize.Width,
						contentBounds.Y+topSize.Height,
					),
					Size: centerSize,
				})
			}
		}
	}

	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *frameLayout) calculateIdealSize(element *ui.Element) ui.Size {
	var (
		leftSize   ui.Size
		rightSize  ui.Size
		topSize    ui.Size
		bottomSize ui.Size
		centerSize ui.Size
	)

	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}

		switch layoutConfig.VerticalAlignment {
		case VerticalAlignmentTop:
			topSize = ui.Size{
				Width:  maxInt(topSize.Width, childSize.Width),
				Height: maxInt(topSize.Height, childSize.Height),
			}
		case VerticalAlignmentBottom:
			bottomSize = ui.Size{
				Width:  maxInt(bottomSize.Width, childSize.Width),
				Height: maxInt(bottomSize.Height, childSize.Height),
			}
		default: // treat as center
			switch layoutConfig.HorizontalAlignment {
			case HorizontalAlignmentLeft:
				leftSize = ui.Size{
					Width:  maxInt(leftSize.Width, childSize.Width),
					Height: maxInt(leftSize.Height, childSize.Height),
				}
			case HorizontalAlignmentRight:
				rightSize = ui.Size{
					Width:  maxInt(rightSize.Width, childSize.Width),
					Height: maxInt(rightSize.Height, childSize.Height),
				}
			default: // treat as center
				centerSize = ui.Size{
					Width:  maxInt(centerSize.Width, childSize.Width),
					Height: maxInt(centerSize.Height, childSize.Height),
				}
			}
		}
	}

	result := ui.Size{
		Width: maxInt(
			maxInt(topSize.Width, bottomSize.Width),
			leftSize.Width+centerSize.Width+rightSize.Width,
		),
		Height: topSize.Height + bottomSize.Height + maxInt(
			maxInt(leftSize.Height, rightSize.Height),
			centerSize.Height,
		),
	}
	return result.Grow(element.Padding().Size())
}
