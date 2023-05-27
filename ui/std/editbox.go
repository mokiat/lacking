package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	EditboxMinWidth = 100
	EditboxHeight   = 28
	EditboxFontFile = "ui:///roboto-regular.ttf"
	EditboxFontSize = float32(18)
)

// EditboxData holds the data for the Editbox component.
type EditboxData struct {

	// Text specifies the committed text in the Editbox.
	Text string
}

var editboxDefaultData = EditboxData{}

// EditboxCallbackData holds the callback data for the Editbox component.
type EditboxCallbackData struct {

	// OnChanged is called when a new string is confirmed by the user.
	OnChanged func(text string)
}

var editboxDefaultCallbackData = EditboxCallbackData{
	OnChanged: func(text string) {},
}

// Editbox is a component that allows a user to input a string.
var Editbox = co.Define(&EditboxComponent{})

type EditboxComponent struct {
	Scope      co.Scope      `co:"scope"`
	Properties co.Properties `co:"properties"`

	font         *ui.Font
	textSize     sprec.Vec2
	text         string
	volatileText string

	onChanged func(string)
}

func (c *EditboxComponent) OnUpsert() {
	c.font = co.OpenFont(c.Scope, EditboxFontFile)

	data := co.GetOptionalData(c.Properties, editboxDefaultData)
	if data.Text != c.text {
		c.text = data.Text
		c.volatileText = data.Text
	}
	c.textSize = c.font.TextSize(c.text, EditboxFontSize)

	callbackData := co.GetOptionalCallbackData(c.Properties, editboxDefaultCallbackData)
	c.onChanged = callbackData.OnChanged
}

func (c *EditboxComponent) Render() co.Instance {
	// force specific height
	layoutData := co.GetOptionalLayoutData(c.Properties, layout.Data{})
	layoutData.Height = opt.V(EditboxHeight)

	return co.New(co.Element, func() {
		co.WithLayoutData(layoutData)
		co.WithData(co.ElementData{
			Essence:   c,
			Focusable: opt.V(true),
			IdealSize: opt.V(ui.NewSize(
				EditboxMinWidth,
				EditboxHeight,
			)),
		})
	})
}

func (c *EditboxComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	switch event.Type {
	case ui.KeyboardEventTypeKeyDown, ui.KeyboardEventTypeRepeat:
		switch event.Code {
		case ui.KeyCodeBackspace:
			if len(c.volatileText) > 0 {
				c.volatileText = c.volatileText[:len(c.volatileText)-1]
				element.Invalidate()
			}
		case ui.KeyCodeEscape:
			c.volatileText = c.text
			element.Window().DiscardFocus()
		case ui.KeyCodeEnter:
			c.text = c.volatileText
			c.onChanged(c.text)
			element.Window().DiscardFocus()
		}
	case ui.KeyboardEventTypeType:
		c.volatileText += string(event.Rune)
		element.Invalidate()
	}
	return true
}

func (c *EditboxComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var strokeColor ui.Color
	var text string
	if element.Window().IsElementFocused(element) {
		strokeColor = SecondaryColor
		text = c.volatileText + "|"
	} else {
		strokeColor = PrimaryLightColor
		text = c.volatileText
	}

	size := element.Bounds().Size
	width := float32(size.Width)
	height := float32(size.Height)

	canvas.Reset()
	canvas.SetStrokeSize(2.0)
	canvas.SetStrokeColor(strokeColor)
	canvas.RoundRectangle(
		sprec.ZeroVec2(),
		sprec.NewVec2(width, height),
		sprec.NewVec4(8, 8, 8, 8),
	)
	canvas.Fill(ui.Fill{
		Color: SurfaceColor,
	})
	canvas.Stroke()

	textPosition := sprec.NewVec2(5, (height-c.textSize.Y)/2)
	canvas.Push()
	canvas.SetClipRect(5, width-5, 2, height-2)
	canvas.Reset()
	canvas.FillText(text, textPosition, ui.Typography{
		Font:  c.font,
		Size:  EditboxFontSize,
		Color: OnSurfaceColor,
	})
	canvas.Pop()
}
