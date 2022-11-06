package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

var (
	EditboxMinWidth = 100
	EditboxHeight   = 28
	EditboxFontFile = "mat:///roboto-regular.ttf"
	EditboxFontSize = float32(18)
)

// EditboxData holds the data for the Editbox component.
type EditboxData struct {

	// Text specifies the committed text in the Editbox.
	Text string
}

var defaultEditboxData = EditboxData{}

// EditboxCallbackData holds the callback data for the Editbox component.
type EditboxCallbackData struct {

	// OnChanged is called when a new string is confirmed by the user.
	OnChanged func(text string)
}

var defaultEditboxCallbackData = EditboxCallbackData{
	OnChanged: func(text string) {},
}

// Editbox is a component that allows a user to input a string.
var Editbox = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data         = co.GetOptionalData(props, defaultEditboxData)
		layoutData   = co.GetOptionalLayoutData(props, LayoutData{})
		callbackData = co.GetOptionalCallbackData(props, defaultEditboxCallbackData)
	)

	essence := co.UseState(func() *editboxEssence {
		return &editboxEssence{}
	}).Get()
	essence.font = co.OpenFont(scope, EditboxFontFile)
	essence.onChanged = callbackData.OnChanged
	if data.Text != essence.text {
		essence.text = data.Text
		essence.volatileText = data.Text
	}
	essence.textSize = essence.font.TextSize(data.Text, EditboxFontSize)

	// force specific height
	layoutData.Height = opt.V(EditboxHeight)

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence:   essence,
			Focusable: opt.V(true),
			IdealSize: opt.V(ui.NewSize(
				EditboxMinWidth,
				EditboxHeight,
			)),
		})
		co.WithLayoutData(layoutData)
	})
})

var _ ui.ElementKeyboardHandler = (*editboxEssence)(nil)
var _ ui.ElementRenderHandler = (*editboxEssence)(nil)

type editboxEssence struct {
	text         string
	textSize     sprec.Vec2
	volatileText string
	font         *ui.Font
	onChanged    func(string)
}

func (e *editboxEssence) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	switch event.Type {
	case ui.KeyboardEventTypeKeyDown, ui.KeyboardEventTypeRepeat:
		switch event.Code {
		case ui.KeyCodeBackspace:
			if len(e.volatileText) > 0 {
				e.volatileText = e.volatileText[:len(e.volatileText)-1]
				element.Invalidate()
			}
		case ui.KeyCodeEscape:
			e.volatileText = e.text
			element.Window().DiscardFocus()
		case ui.KeyCodeEnter:
			e.text = e.volatileText
			e.onChanged(e.text)
			element.Window().DiscardFocus()
		}
	case ui.KeyboardEventTypeType:
		e.volatileText += string(event.Rune)
		element.Invalidate()
	}
	return true
}

func (e *editboxEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var strokeColor ui.Color
	var text string
	if element.Window().IsElementFocused(element) {
		strokeColor = SecondaryColor
		text = e.volatileText + "|"
	} else {
		strokeColor = PrimaryLightColor
		text = e.volatileText
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

	textPosition := sprec.NewVec2(5, (height-e.textSize.Y)/2)
	canvas.Push()
	canvas.SetClipRect(5, width-5, 2, height-2)
	canvas.Reset()
	canvas.FillText(text, textPosition, ui.Typography{
		Font:  e.font,
		Size:  EditboxFontSize,
		Color: OnSurfaceColor,
	})
	canvas.Pop()
}
