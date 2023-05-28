package std

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
	ButtonFontFile       = "ui:///roboto-regular.ttf"
	ButtonFontSize       = float32(18)
)

// OnActionFunc can be used to get notifications about action (click) events.
type OnActionFunc func()

// ButtonData holds the data for the Button component.
type ButtonData struct {
	Text    string
	Icon    *ui.Image
	Enabled opt.T[bool]
}

var buttonDefaultData = ButtonData{}

// ButtonCallbackData holds the callback data for the Button component.
type ButtonCallbackData struct {
	OnClick OnActionFunc
}

var buttonDefaultCallbackData = ButtonCallbackData{
	OnClick: func() {},
}

// Button is a component that allows a user click on it to activate a process.
var Button = co.Define(&buttonComponent{})

type buttonComponent struct {
	co.BaseComponent
	BaseButtonComponent

	icon      *ui.Image
	text      string
	isEnabled bool
}

func (c *buttonComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), buttonDefaultData)
	c.icon = data.Icon
	c.text = data.Text
	c.isEnabled = !data.Enabled.Specified || data.Enabled.Value

	callbackData := co.GetOptionalCallbackData(c.Properties(), buttonDefaultCallbackData)
	c.SetOnClickFunc(callbackData.OnClick)
}

func (c *buttonComponent) Render() co.Instance {
	// force specific height
	layoutData := co.GetOptionalLayoutData(c.Properties(), layout.Data{})
	layoutData.Height = opt.V(ButtonHeight)

	foregroundColor := OnSurfaceColor
	if !c.isEnabled {
		foregroundColor = OutlineColor
	}

	return co.New(co.Element, func() {
		co.WithLayoutData(layoutData)
		co.WithData(co.ElementData{
			Essence: c,
			Layout: layout.Horizontal(layout.HorizontalSettings{
				ContentAlignment: layout.VerticalAlignmentCenter,
				ContentSpacing:   ButtonContentSpacing,
			}),
			Padding: ui.Spacing{
				Left:  ButtonSidePadding,
				Right: ButtonSidePadding,
			},
			Enabled: opt.V(c.isEnabled),
		})

		if c.icon != nil {
			co.WithChild("icon", co.New(Picture, func() {
				co.WithData(PictureData{
					Image:      c.icon,
					ImageColor: opt.V(foregroundColor),
					Mode:       ImageModeFit,
				})
				co.WithLayoutData(layout.Data{
					Width:  opt.V(ButtonIconSize),
					Height: opt.V(ButtonIconSize),
				})
			}))
		}

		if c.text != "" {
			co.WithChild("text", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(c.Scope(), ButtonFontFile),
					FontSize:  opt.V(float32(ButtonFontSize)),
					FontColor: opt.V(foregroundColor),
					Text:      c.text,
				})
				co.WithLayoutData(layout.Data{})
			}))
		}
	})
}

func (c *buttonComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	backgroundColor := SurfaceColor
	strokeColor := PrimaryLightColor
	if element.Enabled() {
		switch c.State() {
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

// ButtonState indicates the state of a Button control.
type ButtonState int

const (
	// ButtonStateUp indicates that the button is in its default state.
	ButtonStateUp ButtonState = iota

	// ButtonStateOver indicates that the cursor is over the button.
	ButtonStateOver

	// ButtonStateDown indicates that the cursor is pressing on the button.
	ButtonStateDown
)

// BaseButtonComponent provides a basic mouse event handling for
// a button component.
//
// Users are expected to compose this structure into a component implementation
// that can do the actual rendering.
type BaseButtonComponent struct {
	state   ButtonState
	onClick OnActionFunc
}

func (c *BaseButtonComponent) OnClickFunc() OnActionFunc {
	return c.onClick
}

func (c *BaseButtonComponent) SetOnClickFunc(onAction OnActionFunc) {
	c.onClick = onAction
}

func (c *BaseButtonComponent) State() ButtonState {
	return c.state
}

func (c *BaseButtonComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Type {
	case ui.MouseEventTypeEnter:
		c.state = ButtonStateOver
		element.Invalidate()
		return true
	case ui.MouseEventTypeLeave:
		c.state = ButtonStateUp
		element.Invalidate()
		return true
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if c.state == ButtonStateDown {
				c.notifyClicked()
			}
			c.state = ButtonStateOver
			element.Invalidate()
			return true
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			c.state = ButtonStateDown
			element.Invalidate()
			return true
		}
	}
	return false
}

func (c *BaseButtonComponent) notifyClicked() {
	if c.onClick != nil {
		c.onClick()
	}
}
