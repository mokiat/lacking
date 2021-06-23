package standard

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterControlBuilder("View", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildView(ctx, template, layoutConfig)
	}))
}

// View represents a control that is the basis of a view.
type View interface {
	ui.Control

	// AddControl adds a control to this view.
	AddControl(control ui.Control)

	// RemoveControl removes a control from this view.
	RemoveControl(control ui.Control)
}

// BuildView constructs a new View control.
func BuildView(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (View, error) {
	result := &view{}

	element := ctx.CreateElement()
	element.SetLayoutConfig(layoutConfig)
	element.SetHandler(result)

	result.Control = ctx.CreateControl(element)
	element.SetControl(result)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}

	for _, childTemplate := range template.Children() {
		child, err := ctx.InstantiateTemplate(childTemplate, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate child from template: %w", err)
		}
		result.AddControl(child)
	}

	return result, nil
}

var _ ui.ElementResizeHandler = (*view)(nil)
var _ ui.ElementRenderHandler = (*view)(nil)

type view struct {
	ui.Control

	backgroundColor *ui.Color
}

func (v *view) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := v.Element().ApplyAttributes(attributes); err != nil {
		return err
	}
	context := v.Element().Context()
	if colorValue, ok := attributes.ColorAttribute("background-color"); ok {
		v.backgroundColor = &colorValue
	}
	if stringValue, ok := attributes.StringAttribute("font"); ok {
		if _, err := context.OpenFontCollection(stringValue); err != nil {
			return fmt.Errorf("failed to load font collection: %w", err)
		}
	}
	return nil
}

func (v *view) AddControl(control ui.Control) {
	v.Element().AppendChild(control.Element())
}

func (v *view) RemoveControl(control ui.Control) {
	v.Element().RemoveChild(control.Element())
}

func (v *view) OnResize(element *ui.Element, bounds ui.Bounds) {
	contentBounds := v.Element().ContentBounds()
	for childElement := v.Element().FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		childElement.SetBounds(contentBounds)
	}
}

func (v *view) OnRender(element *ui.Element, canvas ui.Canvas) {
	if v.backgroundColor != nil {
		canvas.SetSolidColor(*v.backgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			v.Element().Bounds().Size,
		)
	}
}
