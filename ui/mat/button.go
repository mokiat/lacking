package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	ButtonSidePadding    = 6
	ButtonContentSpacing = 5
	ButtonIconSize       = 24
	ButtonHeight         = 28
	ButtonFontFile       = "mat:///roboto-regular.ttf"
	ButtonFontSize       = float32(18)
)

// ButtonData holds the data for the Button component.
type ButtonData struct {
	Text    string
	Icon    *ui.Image
	Enabled opt.T[bool]
}

var defaultButtonData = ButtonData{}

// ButtonCallbackData holds the callback data for the Button component.
type ButtonCallbackData struct {
	ClickListener ClickListener
}

var defaultButtonCallbackData = ButtonCallbackData{
	ClickListener: func() {},
}

// Button is a component that allows a user click on it to activate a process.
var Button = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data         = co.GetOptionalData(props, defaultButtonData)
		layoutData   = co.GetOptionalLayoutData(props, layout.Data{})
		callbackData = co.GetOptionalCallbackData(props, defaultButtonCallbackData)
	)

	essence := co.UseState(func() *buttonEssence {
		return &buttonEssence{
			ButtonBaseEssence: NewButtonBaseEssence(callbackData.ClickListener),
		}
	}).Get()

	// force specific height
	layoutData.Height = opt.V(ButtonHeight)

	foregroundColor := OnSurfaceColor
	if data.Enabled.Specified && !data.Enabled.Value {
		foregroundColor = OutlineColor
	}

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Layout: layout.Horizontal(layout.HorizontalSettings{
				ContentAlignment: layout.VerticalAlignmentCenter,
				ContentSpacing:   ButtonContentSpacing,
			}),
			Padding: ui.Spacing{
				Left:  ButtonSidePadding,
				Right: ButtonSidePadding,
			},
			Enabled: data.Enabled,
		})
		co.WithLayoutData(layoutData)

		if data.Icon != nil {
			co.WithChild("icon", co.New(Picture, func() {
				co.WithData(PictureData{
					Image:      data.Icon,
					ImageColor: opt.V(foregroundColor),
					Mode:       ImageModeFit,
				})
				co.WithLayoutData(layout.Data{
					Width:  opt.V(ButtonIconSize),
					Height: opt.V(ButtonIconSize),
				})
			}))
		}

		if data.Text != "" {
			co.WithChild("text", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(scope, ButtonFontFile),
					FontSize:  opt.V(float32(ButtonFontSize)),
					FontColor: opt.V(foregroundColor),
					Text:      data.Text,
				})
				co.WithLayoutData(layout.Data{})
			}))
		}
	})
})

var _ ui.ElementRenderHandler = (*buttonEssence)(nil)

type buttonEssence struct {
	*ButtonBaseEssence
}

func (e *buttonEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	backgroundColor := SurfaceColor
	strokeColor := PrimaryLightColor
	if element.Enabled() {
		switch e.State() {
		case ButtonStateOver:
			backgroundColor = backgroundColor.Overlay(HoverOverlayColor)
		case ButtonStateDown:
			backgroundColor = backgroundColor.Overlay(PressOverlayColor)
		}
	} else {
		strokeColor = OutlineColor
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
		Color: backgroundColor,
	})
	canvas.Stroke()
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

// SetOnClick changes the ClickListener.
func (e *ButtonBaseEssence) SetOnClick(onClick ClickListener) {
	e.onClick = onClick
}

// State returns the current state of the button.
func (e *ButtonBaseEssence) State() ButtonState {
	return e.state
}

func (e *ButtonBaseEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Type {
	case ui.MouseEventTypeEnter:
		e.state = ButtonStateOver
		element.Invalidate()
		return true
	case ui.MouseEventTypeLeave:
		e.state = ButtonStateUp
		element.Invalidate()
		return true
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if e.state == ButtonStateDown {
				e.onClick()
			}
			e.state = ButtonStateOver
			element.Invalidate()
			return true
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			e.state = ButtonStateDown
			element.Invalidate()
			return true
		}
	}
	return false
}
