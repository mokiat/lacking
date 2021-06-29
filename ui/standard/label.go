package standard

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

// Label represents a Control that displays an immutable
// text string.
type Label interface {
	ui.Control
}

func init() {
	ui.RegisterControlBuilder("Label", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildLabel(ctx, template, layoutConfig)
	}))
}

// BuildLabel constructs a new Label control.
func BuildLabel(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (Label, error) {
	result := &label{
		element:   ctx.CreateElement(),
		textSize:  24,
		textColor: ui.White(),
	}
	result.element.SetLayoutConfig(layoutConfig)
	result.element.SetEssence(result)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}
	return result, nil
}

var _ ui.ElementRenderHandler = (*label)(nil)

type label struct {
	element *ui.Element

	font      ui.Font
	text      string
	textSize  int
	textColor ui.Color
}

func (b *label) Element() *ui.Element {
	return b.element
}

func (b *label) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := b.element.ApplyAttributes(attributes); err != nil {
		return err
	}
	context := b.element.Context()
	if familyStringValue, ok := attributes.StringAttribute("font-family"); ok {
		if subFamilyStringValue, ok := attributes.StringAttribute("font-style"); ok {
			font, found := context.GetFont(familyStringValue, subFamilyStringValue)
			if !found {
				return fmt.Errorf("could not find font %q / %q", familyStringValue, subFamilyStringValue)
			}
			b.font = font
		}
	}
	if stringValue, ok := attributes.StringAttribute("text"); ok {
		b.text = stringValue
	}
	if intValue, ok := attributes.IntAttribute("text-size"); ok {
		b.textSize = intValue
	}
	if colorValue, ok := attributes.ColorAttribute("text-color"); ok {
		b.textColor = colorValue
	}
	return nil
}

func (b *label) OnRender(element *ui.Element, canvas ui.Canvas) {
	if b.font != nil && b.text != "" {
		canvas.SetFont(b.font)
		canvas.SetFontSize(b.textSize)
		canvas.SetSolidColor(b.textColor)
		canvas.DrawText(b.text, ui.NewPosition(0, 0))
	}
}
