package standard

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterControlBuilder("VerticalLayout", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildVerticalLayout(ctx, template, layoutConfig)
	}))
}

// VerticalLayoutConfig represents a layout configuration for a Control
// that is added to a VerticalLayout.
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

// VerticalLayout represents a layout that positions controls
// vertically in sequence from top to bottom.
type VerticalLayout interface {
	ui.Control

	// AddControl adds a control to this layout.
	AddControl(control ui.Control)

	// RemoveControl removes the specified control from this layout.
	RemoveControl(control ui.Control)
}

// BuildVerticalLayout constructs a new VerticalLayout control.
func BuildVerticalLayout(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (VerticalLayout, error) {
	result := &verticalLayout{}

	element := ctx.CreateElement()
	element.SetLayoutConfig(layoutConfig)
	element.SetHandler(result)

	result.Control = ctx.CreateControl(element)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}

	for _, childTemplate := range template.Children() {
		childLayoutConfig := new(VerticalLayoutConfig)
		childLayoutConfig.ApplyAttributes(childTemplate.LayoutAttributes())
		child, err := ctx.InstantiateTemplate(childTemplate, childLayoutConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate child from template: %w", err)
		}
		result.AddControl(child)
	}

	return result, nil
}

var _ ui.ElementResizeHandler = (*verticalLayout)(nil)
var _ ui.ElementRenderHandler = (*verticalLayout)(nil)

type verticalLayout struct {
	ui.Control

	backgroundColor  *ui.Color
	contentAlignment Alignment
	contentSpacing   int
}

func (l *verticalLayout) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := l.Control.ApplyAttributes(attributes); err != nil {
		return err
	}
	if colorValue, ok := attributes.ColorAttribute("background-color"); ok {
		l.backgroundColor = &colorValue
	}
	if alignmentValue, ok := AlignmentAttribute(attributes, "content-alignment"); ok {
		l.contentAlignment = alignmentValue
	} else {
		l.contentAlignment = AlignmentCenter
	}
	if intValue, ok := attributes.IntAttribute("content-spacing"); ok {
		l.contentSpacing = intValue
	}
	return nil
}

func (l *verticalLayout) AddControl(control ui.Control) {
	l.Element().AppendChild(control.Element())
}

func (l *verticalLayout) RemoveControl(control ui.Control) {
	l.Element().RemoveChild(control.Element())
}

func (l *verticalLayout) OnResize(element *ui.Element, bounds ui.Bounds) {
	contentBounds := l.Element().ContentBounds()

	topPlacement := contentBounds.Y
	for childElement := l.Element().FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
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

func (l *verticalLayout) OnRender(element *ui.Element, canvas ui.Canvas) {
	if l.backgroundColor != nil {
		canvas.SetSolidColor(*l.backgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			l.Element().Bounds().Size,
		)
	}
}
