package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

type ButtonData struct {
	Padding       ui.Spacing
	Font          *ui.Font
	FontSize      optional.V[float32]
	FontColor     optional.V[ui.Color]
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
			state: ButtonStateUp,
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

	txtSize := essence.font.TextSize(essence.text, essence.fontSize)

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Padding: data.Padding,
			IdealSize: optional.Value(
				ui.NewSize(int(txtSize.X), int(txtSize.Y)).Grow(data.Padding.Horizontal(), data.Padding.Vertical()),
			),
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
}))

var _ ui.ElementMouseHandler = (*buttonEssence)(nil)
var _ ui.ElementRenderHandler = (*buttonEssence)(nil)

type buttonEssence struct {
	font          *ui.Font
	fontSize      float32
	fontColor     ui.Color
	fontAlignment Alignment
	text          string

	clickListener ClickListener

	state ButtonState
}

func (e *buttonEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	context := element.Context()
	switch event.Type {
	case ui.MouseEventTypeEnter:
		e.state = ButtonStateOver
		context.Window().Invalidate()
	case ui.MouseEventTypeLeave:
		e.state = ButtonStateUp
		context.Window().Invalidate()
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if e.state == ButtonStateDown {
				e.onClick()
			}
			e.state = ButtonStateOver
			context.Window().Invalidate()
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			e.state = ButtonStateDown
			context.Window().Invalidate()
		}
	}
	return true
}

func (e *buttonEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	switch e.state {
	case ButtonStateOver:
		canvas.Shape().Begin(ui.Fill{
			Color: ui.RGB(15, 15, 15),
		})
		canvas.Shape().Rectangle(
			sprec.NewVec2(0, 0),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Shape().End()
	case ButtonStateDown:
		canvas.Shape().Begin(ui.Fill{
			Color: ui.RGB(30, 30, 30),
		})
		canvas.Shape().Rectangle(
			sprec.NewVec2(0, 0),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Shape().End()
	}

	if e.font != nil && e.text != "" {
		canvas.Text().Begin(ui.Typography{
			Font:  e.font,
			Size:  e.fontSize,
			Color: e.fontColor,
		})
		var textPosition sprec.Vec2
		contentArea := element.ContentBounds()
		textDrawSize := e.font.TextSize(e.text, e.fontSize)
		switch e.fontAlignment {
		case AlignmentLeft:
			textPosition = sprec.NewVec2(
				float32(contentArea.X),
				float32(contentArea.Y),
			)
		default:
			textPosition = sprec.NewVec2(
				float32(contentArea.X)+(float32(contentArea.Width)-textDrawSize.X)/2,
				float32(contentArea.Y)+(float32(contentArea.Height)-textDrawSize.Y)/2,
			)
		}
		canvas.Text().Line(e.text, textPosition)
		canvas.Text().End()
	}
}

func (e *buttonEssence) onClick() {
	if e.clickListener != nil {
		e.clickListener()
	}
}
