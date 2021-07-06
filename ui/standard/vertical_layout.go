package standard

import (
	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterLayoutBuilder("VerticalLayout", ui.LayoutBuilderFunc(func(attributes ui.AttributeSet) (ui.Layout, error) {
		return NewVerticalLayout(attributes), nil
	}))
}

// NewVerticalLayout creates a new vertical layout instance.
func NewVerticalLayout(attributes ui.AttributeSet) *VerticalLayout {
	result := &VerticalLayout{}
	if alignmentValue, ok := AlignmentAttribute(attributes, "content-alignment"); ok {
		result.contentAlignment = alignmentValue
	} else {
		result.contentAlignment = AlignmentCenter
	}
	if intValue, ok := attributes.IntAttribute("content-spacing"); ok {
		result.contentSpacing = intValue
	}
	return result
}

var _ ui.Layout = (*VerticalLayout)(nil)

type VerticalLayout struct {
	contentAlignment Alignment
	contentSpacing   int
}

// LayoutConfig creates a new layout config instance specific
// to this layout.
func (l *VerticalLayout) LayoutConfig() ui.LayoutConfig {
	return &VerticalLayoutConfig{}
}

// Apply applies this layout to the specified Element.
func (l *VerticalLayout) Apply(element *ui.Element) {
	contentBounds := element.ContentBounds()

	topPlacement := contentBounds.Y
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := childElement.LayoutConfig().(*VerticalLayoutConfig)

		childBounds := ui.Bounds{
			Size: childElement.IdealSize(),
		}
		if layoutConfig.Width != nil {
			childBounds.Width = *layoutConfig.Width
		}
		if layoutConfig.Height != nil {
			childBounds.Height = *layoutConfig.Height
		}

		switch l.contentAlignment {
		case AlignmentLeft:
			childBounds.X = contentBounds.X + childElement.Margin().Left
		case AlignmentRight:
			childBounds.X = contentBounds.X + contentBounds.Width - childElement.Margin().Right - childBounds.Width
		case AlignmentCenter:
			fallthrough
		default:
			childBounds.X = contentBounds.X + (contentBounds.Width-childBounds.Width)/2 - +childElement.Margin().Left
		}

		childBounds.Y = topPlacement + childElement.Margin().Top
		childElement.SetBounds(childBounds)

		topPlacement += childElement.Margin().Vertical() + childBounds.Height + l.contentSpacing
	}
}

// VerticalLayoutConfig represents a layout configuration for a Control
// that is added to a VerticalLayoutLayout.
type VerticalLayoutConfig struct {
	Width  *int
	Height *int
}

func (c *VerticalLayoutConfig) ApplyAttributes(attributes ui.AttributeSet) {
	if width, ok := attributes.IntAttribute("width"); ok {
		c.Width = &width
	}
	if height, ok := attributes.IntAttribute("height"); ok {
		c.Height = &height
	}
}
