package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	ToolbarButtonSidePadding     = 4
	ToolbarButtonContentSpacing  = 5
	ToolbarButtonIconSize        = 24
	ToolbarButtonFontFile        = "ui:///roboto-regular.ttf"
	ToolbarButtonFontSize        = float32(20)
	ToolbarBottonSelectionHeight = float32(5.0)
)

// ToolbarButtonData holds the data for a ToolbarButton component.
type ToolbarButtonData struct {
	Icon     *ui.Image
	Text     string
	Enabled  opt.T[bool]
	Selected bool
}

var toolbarButtonDefaultData = ToolbarButtonData{}

// ToolbarButtonCallbackData holds the callback handlers for a
// ToolbarButton component.
type ToolbarButtonCallbackData struct {
	OnClick OnActionFunc
}

var toolbarButtonDefaultCallbackData = ToolbarButtonCallbackData{
	OnClick: func() {},
}

var ToolbarButton = co.Define(&toolbarButtonComponent{})

type toolbarButtonComponent struct {
	co.BaseComponent
	BaseButtonComponent

	icon       *ui.Image
	text       string
	isEnabled  bool
	isSelected bool
}

func (c *toolbarButtonComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), toolbarButtonDefaultData)
	c.icon = data.Icon
	c.text = data.Text
	c.isEnabled = !data.Enabled.Specified || data.Enabled.Value
	c.isSelected = data.Selected

	callbackData := co.GetOptionalCallbackData(c.Properties(), toolbarButtonDefaultCallbackData)
	c.SetOnClickFunc(callbackData.OnClick)
}

func (c *toolbarButtonComponent) Render() co.Instance {
	// force specific height
	layoutData := co.GetOptionalLayoutData(c.Properties(), layout.Data{})
	layoutData.Height = opt.V(ToolbarItemHeight)

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
				ContentSpacing:   ToolbarButtonContentSpacing,
			}),
			Padding: ui.Spacing{
				Left:  ToolbarButtonSidePadding,
				Right: ToolbarButtonSidePadding,
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
					Width:  opt.V(ToolbarButtonIconSize),
					Height: opt.V(ToolbarButtonIconSize),
				})
			}))
		}

		if c.text != "" {
			co.WithChild("text", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(c.Scope(), ToolbarButtonFontFile),
					FontSize:  opt.V(float32(ToolbarButtonFontSize)),
					FontColor: opt.V(foregroundColor),
					Text:      c.text,
				})
				co.WithLayoutData(layout.Data{})
			}))
		}
	})
}

func (c *toolbarButtonComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch c.State() {
	case ButtonStateOver:
		backgroundColor = HoverOverlayColor
	case ButtonStateDown:
		backgroundColor = PressOverlayColor
	default:
		backgroundColor = ui.Transparent()
	}

	drawBounds := canvas.DrawBounds(element, false)

	if !backgroundColor.Transparent() {
		canvas.Reset()
		canvas.Rectangle(
			drawBounds.Position,
			drawBounds.Size,
		)
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
	if c.isSelected {
		canvas.Reset()
		canvas.Rectangle(
			sprec.NewVec2(0.0, drawBounds.Height()-ToolbarBottonSelectionHeight),
			sprec.NewVec2(drawBounds.Width(), ToolbarBottonSelectionHeight),
		)
		canvas.Fill(ui.Fill{
			Color: SecondaryColor,
		})
	}
}
