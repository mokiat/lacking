package layout

import "github.com/mokiat/lacking/ui"

// IDEA: Consider using `Left`, `Right`, `Top`, `Bottom` as margins for
// elements positioned by a horizontal layout.

// HorizontalSettings contains configurations for the Horizontal layout.
type HorizontalSettings struct {
	ContentAlignment VerticalAlignment
	ContentSpacing   int
}

// Horizontal returns a layout that positions elements horizontally,
// in a sequence.
func Horizontal(settings ...HorizontalSettings) ui.Layout {
	cfg := HorizontalSettings{}
	if len(settings) > 0 {
		cfg = settings[0]
	}
	return &horizontalLayout{
		contentAlignment: cfg.ContentAlignment,
		contentSpacing:   cfg.ContentSpacing,
	}
}

type horizontalLayout struct {
	contentAlignment VerticalAlignment
	contentSpacing   int
}

func (l *horizontalLayout) Apply(element *ui.Element) {
	l.applyLeftToRight(element)
	l.applyRightToLeft(element)
	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *horizontalLayout) applyLeftToRight(element *ui.Element) {
	contentBounds := element.ContentBounds()

	leftPlacement := contentBounds.X
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)
		if layoutConfig.HorizontalAlignment == HorizontalAlignmentRight {
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
		if layoutConfig.GrowVertically {
			childBounds.Height = contentBounds.Height
		}

		var alignment VerticalAlignment
		switch layoutConfig.VerticalAlignment {
		case VerticalAlignmentTop:
			alignment = VerticalAlignmentTop
		case VerticalAlignmentBottom:
			alignment = VerticalAlignmentBottom
		case VerticalAlignmentCenter:
			alignment = VerticalAlignmentCenter
		case VerticalAlignmentDefault:
			fallthrough
		default:
			switch l.contentAlignment {
			case VerticalAlignmentTop:
				alignment = VerticalAlignmentTop
			case VerticalAlignmentBottom:
				alignment = VerticalAlignmentBottom
			case VerticalAlignmentCenter:
				fallthrough
			case VerticalAlignmentDefault:
				fallthrough
			default:
				alignment = VerticalAlignmentCenter
			}
		}

		switch alignment {
		case VerticalAlignmentTop:
			childBounds.Y = contentBounds.Y
		case VerticalAlignmentBottom:
			childBounds.Y = contentBounds.Y + contentBounds.Height - childBounds.Height
		case VerticalAlignmentCenter:
			fallthrough
		case VerticalAlignmentDefault:
			fallthrough
		default:
			childBounds.Y = contentBounds.Y + (contentBounds.Height-childBounds.Height)/2
		}

		childBounds.X = leftPlacement
		childElement.SetBounds(childBounds)

		leftPlacement += childBounds.Width + l.contentSpacing
	}
}

func (l *horizontalLayout) applyRightToLeft(element *ui.Element) {
	contentBounds := element.ContentBounds()

	rightPlacement := contentBounds.Width
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)
		if layoutConfig.HorizontalAlignment != HorizontalAlignmentRight {
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
		if layoutConfig.GrowVertically {
			childBounds.Height = contentBounds.Height
		}

		var alignment VerticalAlignment
		switch layoutConfig.VerticalAlignment {
		case VerticalAlignmentTop:
			alignment = VerticalAlignmentTop
		case VerticalAlignmentBottom:
			alignment = VerticalAlignmentBottom
		case VerticalAlignmentCenter:
			alignment = VerticalAlignmentCenter
		case VerticalAlignmentDefault:
			fallthrough
		default:
			switch l.contentAlignment {
			case VerticalAlignmentTop:
				alignment = VerticalAlignmentTop
			case VerticalAlignmentBottom:
				alignment = VerticalAlignmentBottom
			case VerticalAlignmentCenter:
				fallthrough
			case VerticalAlignmentDefault:
				fallthrough
			default:
				alignment = VerticalAlignmentCenter
			}
		}

		switch alignment {
		case VerticalAlignmentTop:
			childBounds.Y = contentBounds.Y
		case VerticalAlignmentBottom:
			childBounds.Y = contentBounds.Y + contentBounds.Height - childBounds.Height
		case VerticalAlignmentCenter:
			fallthrough
		case VerticalAlignmentDefault:
			fallthrough
		default:
			childBounds.Y = contentBounds.Y + (contentBounds.Height-childBounds.Height)/2
		}

		childBounds.X = rightPlacement - childBounds.Width
		childElement.SetBounds(childBounds)

		rightPlacement -= childBounds.Width + l.contentSpacing
	}
}

func (l *horizontalLayout) calculateIdealSize(element *ui.Element) ui.Size {
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

		result.Height = maxInt(result.Height, childSize.Height)
		result.Width += childSize.Width + l.contentSpacing
	}
	if result.Width > 0 {
		result.Width -= l.contentSpacing // remove last spacing
	}
	return result.Grow(element.Padding().Size())
}
