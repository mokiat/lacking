package layout

import "github.com/mokiat/lacking/ui"

// IDEA: Consider using `Left`, `Right`, `Top`, `Bottom` as margins for
// elements positioned by a vertical layout.

// VerticalSettings contains configurations for the Vertical layout.
type VerticalSettings struct {
	ContentAlignment HorizontalAlignment
	ContentSpacing   int
}

// Vertical returns a layout that positions elements vertically, in a sequence.
func Vertical(settings ...VerticalSettings) ui.Layout {
	cfg := VerticalSettings{}
	if len(settings) > 0 {
		cfg = settings[0]
	}
	return &verticalLayout{
		contentAlignment: cfg.ContentAlignment,
		contentSpacing:   cfg.ContentSpacing,
	}
}

type verticalLayout struct {
	contentAlignment HorizontalAlignment
	contentSpacing   int
}

func (l *verticalLayout) Apply(element *ui.Element) {
	l.applyTopToBottom(element)
	l.applyBottomToTop(element)
	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *verticalLayout) applyTopToBottom(element *ui.Element) {
	contentBounds := element.ContentBounds()

	topPlacement := contentBounds.Y
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)
		if layoutConfig.VerticalAlignment == VerticalAlignmentBottom {
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

		var alignment HorizontalAlignment
		switch layoutConfig.HorizontalAlignment {
		case HorizontalAlignmentLeft:
			alignment = HorizontalAlignmentLeft
		case HorizontalAlignmentRight:
			alignment = HorizontalAlignmentRight
		case HorizontalAlignmentCenter:
			alignment = HorizontalAlignmentCenter
		case HorizontalAlignmentDefault:
			fallthrough
		default:
			switch l.contentAlignment {
			case HorizontalAlignmentLeft:
				alignment = HorizontalAlignmentLeft
			case HorizontalAlignmentRight:
				alignment = HorizontalAlignmentRight
			case HorizontalAlignmentCenter:
				fallthrough
			case HorizontalAlignmentDefault:
				fallthrough
			default:
				alignment = HorizontalAlignmentCenter
			}
		}

		switch alignment {
		case HorizontalAlignmentLeft:
			childBounds.X = contentBounds.X
		case HorizontalAlignmentRight:
			childBounds.X = contentBounds.X + contentBounds.Width - childBounds.Width
		case HorizontalAlignmentCenter:
			fallthrough
		case HorizontalAlignmentDefault:
			fallthrough
		default:
			childBounds.X = contentBounds.X + (contentBounds.Width-childBounds.Width)/2
		}

		childBounds.Y = topPlacement
		childElement.SetBounds(childBounds)

		topPlacement += childBounds.Height + l.contentSpacing
	}
}

func (l *verticalLayout) applyBottomToTop(element *ui.Element) {
	contentBounds := element.ContentBounds()

	bottomPlacement := contentBounds.Height
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)
		if layoutConfig.VerticalAlignment != VerticalAlignmentBottom {
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

		var alignment HorizontalAlignment
		switch layoutConfig.HorizontalAlignment {
		case HorizontalAlignmentLeft:
			alignment = HorizontalAlignmentLeft
		case HorizontalAlignmentRight:
			alignment = HorizontalAlignmentRight
		case HorizontalAlignmentCenter:
			alignment = HorizontalAlignmentCenter
		case HorizontalAlignmentDefault:
			fallthrough
		default:
			switch l.contentAlignment {
			case HorizontalAlignmentLeft:
				alignment = HorizontalAlignmentLeft
			case HorizontalAlignmentRight:
				alignment = HorizontalAlignmentRight
			case HorizontalAlignmentCenter:
				fallthrough
			case HorizontalAlignmentDefault:
				fallthrough
			default:
				alignment = HorizontalAlignmentCenter
			}
		}

		switch alignment {
		case HorizontalAlignmentLeft:
			childBounds.X = contentBounds.X
		case HorizontalAlignmentRight:
			childBounds.X = contentBounds.X + contentBounds.Width - childBounds.Width
		case HorizontalAlignmentCenter:
			fallthrough
		case HorizontalAlignmentDefault:
			fallthrough
		default:
			childBounds.X = contentBounds.X + (contentBounds.Width-childBounds.Width)/2
		}

		childBounds.Y = bottomPlacement - childBounds.Height
		childElement.SetBounds(childBounds)

		bottomPlacement -= childBounds.Height + l.contentSpacing
	}
}

func (l *verticalLayout) calculateIdealSize(element *ui.Element) ui.Size {
	result := ui.NewSize(0, 0)
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}

		result.Width = maxInt(result.Width, childSize.Width)
		result.Height += childSize.Height + l.contentSpacing
	}
	if result.Height > 0 {
		result.Height -= l.contentSpacing // remove last spacing
	}
	return result.Grow(element.Padding().Size())
}
