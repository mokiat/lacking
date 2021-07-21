package mat

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

type ButtonData struct {
	Font ui.Font
	Text string
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
	props.InjectData(&data)
	props.InjectCallbackData(&callbackData)

	co.UseState(func() interface{} {
		return &buttonEssence{
			state: buttonStateUp,
		}
	}).Inject(&essence)

	essence.font = data.Font
	essence.text = data.Text
	essence.clickListener = callbackData.ClickListener

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
}))

var _ ui.ElementMouseHandler = (*buttonEssence)(nil)
var _ ui.ElementRenderHandler = (*buttonEssence)(nil)

type buttonEssence struct {
	font ui.Font
	text string

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
	canvas.SetSolidColor(ui.RGB(128, 0, 255))
	canvas.FillRectangle(
		ui.NewPosition(0, 0),
		element.Bounds().Size,
	)
}

func (e *buttonEssence) onClick() {
	if e.clickListener != nil {
		e.clickListener()
	}
}
