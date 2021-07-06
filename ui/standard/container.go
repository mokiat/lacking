package standard

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterControlBuilder("Container", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildContainer(ctx, template, layoutConfig)
	}))
}

// Container represents a generic control that holds other controls.
type Container interface {
	ui.Control

	// AddControl adds a control to this layout.
	AddControl(control ui.Control)

	// RemoveControl removes the specified control from this layout.
	RemoveControl(control ui.Control)
}

func BuildContainer(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (Container, error) {
	result := &container{
		element: ctx.CreateElement(),
	}
	result.element.SetLayoutConfig(layoutConfig)
	result.element.SetEssence(result)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}
	for _, childTemplate := range template.Children() {
		childLayoutConfig := result.layout.LayoutConfig()
		childLayoutConfig.ApplyAttributes(childTemplate.LayoutAttributes())
		child, err := ctx.InstantiateTemplate(childTemplate, childLayoutConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate child from template: %w", err)
		}
		result.AddControl(child)
	}
	return result, nil
}

var _ ui.ElementResizeHandler = (*container)(nil)
var _ ui.ElementRenderHandler = (*container)(nil)

type container struct {
	element         *ui.Element
	layout          ui.Layout
	backgroundColor *ui.Color
}

func (l *container) Element() *ui.Element {
	return l.element
}

func (l *container) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := l.element.ApplyAttributes(attributes); err != nil {
		return err
	}
	if layoutName, ok := attributes.StringAttribute("layout"); ok {
		layoutBuilder, ok := ui.NamedLayoutBuilder(layoutName)
		if !ok {
			return fmt.Errorf("unknown layout: %s", layoutName)
		}
		layout, err := layoutBuilder.Build(attributes)
		if err != nil {
			return fmt.Errorf("failed to build layout: %w", err)
		}
		l.layout = layout
	} else {
		return fmt.Errorf("missing layout configuration")
	}
	if color, ok := attributes.ColorAttribute("background-color"); ok {
		l.backgroundColor = &color
	}
	return nil
}

func (l *container) AddControl(control ui.Control) {
	l.element.AppendChild(control.Element())
}

func (l *container) RemoveControl(control ui.Control) {
	l.element.RemoveChild(control.Element())
}

func (l *container) OnResize(element *ui.Element, bounds ui.Bounds) {
	l.layout.Apply(element)
}

func (l *container) OnRender(element *ui.Element, canvas ui.Canvas) {
	if l.backgroundColor != nil {
		canvas.SetSolidColor(*l.backgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			l.element.Bounds().Size,
		)
	}
}
