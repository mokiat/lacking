package mat

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/optional"
)

type ButtonData struct {
	Padding       ui.Spacing
	Font          ui.Font
	FontSize      optional.Int
	FontColor     optional.Color
	FontAlignment Alignment
	Text          string
}

type ButtonCallbackData struct {
	ClickListener ClickListener
}

var Button = co.ShallowCached(co.Define(func(props co.Properties) co.Instance {
	var (
		data         ButtonData
		callbackData ButtonCallbackData
		essence      *buttonEssence
	)
	props.InjectOptionalData(&data, ButtonData{})
	props.InjectOptionalCallbackData(&callbackData, ButtonCallbackData{})

	co.UseState(func() interface{} {
		return &buttonEssence{
			state: buttonStateUp,
		}
	}).Inject(&essence)

	essence.font = data.Font
	if data.FontSize.Specified {
		essence.fontSize = data.FontSize.Value
	} else {
		essence.fontSize = 24
	}
	if data.FontColor.Specified {
		essence.fontColor = data.FontColor.Value
	} else {
		essence.fontColor = ui.Black()
	}
	essence.fontAlignment = data.FontAlignment
	essence.text = data.Text
	essence.clickListener = callbackData.ClickListener

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Padding: data.Padding,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
}))

var _ ui.ElementMouseHandler = (*buttonEssence)(nil)
var _ ui.ElementRenderHandler = (*buttonEssence)(nil)

type buttonEssence struct {
	font          ui.Font
	fontSize      int
	fontColor     ui.Color
	fontAlignment Alignment
	text          string

	clickListener ClickListener

	state buttonState
}

func (e *buttonEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	context := element.Context()
	switch event.Type {
	case ui.MouseEventTypeEnter:
		e.state = buttonStateOver
		context.Window().Invalidate()
	case ui.MouseEventTypeLeave:
		e.state = buttonStateUp
		context.Window().Invalidate()
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if e.state == buttonStateDown {
				e.onClick()
			}
			e.state = buttonStateOver
			context.Window().Invalidate()
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			e.state = buttonStateDown
			context.Window().Invalidate()
		}
	}
	return true
}

func (e *buttonEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	switch e.state {
	case buttonStateOver:
		canvas.SetSolidColor(ui.RGB(15, 15, 15))
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	case buttonStateDown:
		canvas.SetSolidColor(ui.RGB(30, 30, 30))
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	}

	if e.font != nil && e.text != "" {
		canvas.SetFont(e.font)
		canvas.SetFontSize(e.fontSize)
		canvas.SetSolidColor(e.fontColor)

		contentArea := element.ContentBounds()
		textDrawSize := canvas.TextSize(e.text)

		var textPosition ui.Position
		switch e.fontAlignment {
		case AlignmentLeft:
			textPosition = ui.NewPosition(
				contentArea.X,
				contentArea.Y,
			)
		default:
			textPosition = ui.NewPosition(
				contentArea.X+(contentArea.Width-textDrawSize.Width)/2,
				contentArea.Y+(contentArea.Height-textDrawSize.Height)/2,
			)
		}

		canvas.DrawText(e.text, textPosition)
	}
}

func (e *buttonEssence) onClick() {
	if e.clickListener != nil {
		e.clickListener()
	}
}
