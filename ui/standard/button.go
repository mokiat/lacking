package standard

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterControlBuilder("Button", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildButton(ctx, template, layoutConfig)
	}))
}

// Button represents a button UI control.
type Button interface {
	ui.Control

	// SetClickListener configures the specified listener to receive
	// click events. You can call this method with nil to disable
	// any existing listener.
	SetClickListener(listener ButtonClickListener)

	// Click simulates a click event. An event will be sent to
	// the ButtonClickListener, if one is registered.
	Click()
}

// ButtonClickListener can be used to get notifications about
// button click events.
type ButtonClickListener func(button Button)

// BuildButton creates a new Button control.
func BuildButton(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (Button, error) {
	result := &button{
		state: buttonStateUp,
	}

	element := ctx.CreateElement()
	element.SetLayoutConfig(layoutConfig)
	element.SetHandler(result)
	element.SetIdealSize(ui.NewSize(120, 32)) // TODO: Calculate based off of font, label, etc.

	result.Control = ctx.CreateControl(element)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}

	return result, nil
}

var _ ui.ElementRenderHandler = (*button)(nil)
var _ ui.ElementMouseHandler = (*button)(nil)

type button struct {
	ui.Control

	font  ui.Font
	label string

	clickListener ButtonClickListener
	state         buttonState
}

func (b *button) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := b.Control.ApplyAttributes(attributes); err != nil {
		return err
	}
	if stringValue, ok := attributes.StringAttribute("label"); ok {
		b.label = stringValue
	}
	if familyStringValue, ok := attributes.StringAttribute("font-family"); ok {
		if subFamilyStringValue, ok := attributes.StringAttribute("font-style"); ok {
			font, found := b.Context().GetFont(familyStringValue, subFamilyStringValue)
			if !found {
				return fmt.Errorf("could not find font %q / %q", familyStringValue, subFamilyStringValue)
			}
			b.font = font
		}
	}
	return nil
}

func (b *button) SetClickListener(listener ButtonClickListener) {
	b.clickListener = listener
}

func (b *button) Click() {
	if b.clickListener != nil {
		b.clickListener(b)
	}
}

func (b *button) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Type {
	case ui.MouseEventTypeEnter:
		b.state = buttonStateOver
		b.Context().Window().Invalidate()
	case ui.MouseEventTypeLeave:
		b.state = buttonStateUp
		b.Context().Window().Invalidate()
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if b.state == buttonStateDown {
				b.Click()
			}
			b.state = buttonStateOver
			b.Context().Window().Invalidate()
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			b.state = buttonStateDown
			b.Context().Window().Invalidate()
		}
	}
	return true
}

func (b *button) OnRender(element *ui.Element, canvas ui.Canvas) {
	canvas.SetSolidColor(ui.RGB(128, 0, 255))
	canvas.FillRectangle(
		ui.NewPosition(0, 0),
		element.Bounds().Size,
	)
}

type buttonState = int

const (
	buttonStateUp buttonState = 1 + iota
	buttonStateOver
	buttonStateDown
)