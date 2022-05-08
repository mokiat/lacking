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
	)
	props.InjectOptionalData(&data, ButtonData{})
	props.InjectOptionalCallbackData(&callbackData, ButtonCallbackData{})

	essence := co.UseState(func() *buttonEssence {
		return &buttonEssence{
			ButtonBaseEssence: NewButtonBaseEssence(callbackData.ClickListener),
		}
	}).Get()

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

var _ ui.ElementRenderHandler = (*buttonEssence)(nil)

type buttonEssence struct {
	*ButtonBaseEssence
	font          *ui.Font
	fontSize      float32
	fontColor     ui.Color
	fontAlignment Alignment
	text          string
}

func (e *buttonEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	switch e.State() {
	case ButtonStateOver:
		canvas.Reset()
		canvas.Rectangle(
			sprec.NewVec2(0, 0),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Fill(ui.Fill{
			Color: ui.RGB(40, 40, 40),
		})
	case ButtonStateDown:
		canvas.Reset()
		canvas.Rectangle(
			sprec.NewVec2(0, 0),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Fill(ui.Fill{
			Color: ui.RGB(80, 80, 80),
		})
	}

	if e.font != nil && e.text != "" {
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
		canvas.Reset()
		canvas.FillText(e.text, textPosition, ui.Typography{
			Font:  e.font,
			Size:  e.fontSize,
			Color: e.fontColor,
		})
	}
}

// NewButtonBaseEssence creates a new ButtonBaseEssence instance.
func NewButtonBaseEssence(onClick ClickListener) *ButtonBaseEssence {
	return &ButtonBaseEssence{
		state:   ButtonStateUp,
		onClick: onClick,
	}
}

var _ ui.ElementMouseHandler = (*ButtonBaseEssence)(nil)

// ButtonBaseEssence provides a basic mouse event handling for
// a button control.
// You are expected to compose this structure into an essence that
// can do the actual rendering.
type ButtonBaseEssence struct {
	state   ButtonState
	onClick ClickListener
}

func (e *ButtonBaseEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	context := element.Context()
	switch event.Type {
	case ui.MouseEventTypeEnter:
		e.state = ButtonStateOver
		context.Window().Invalidate()
		return true
	case ui.MouseEventTypeLeave:
		e.state = ButtonStateUp
		context.Window().Invalidate()
		return true
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if e.state == ButtonStateDown {
				e.onClick()
			}
			e.state = ButtonStateOver
			context.Window().Invalidate()
			return true
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			e.state = ButtonStateDown
			context.Window().Invalidate()
			return true
		}
	}
	return false
}

// State returns the current state of the button.
func (e *ButtonBaseEssence) State() ButtonState {
	return e.state
}
