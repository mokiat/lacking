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
	ToolbarButtonFontFile        = "mat:///roboto-regular.ttf"
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
	OnClick OnClickFunc
}

var toolbarButtonDefaultCallbackData = ToolbarButtonCallbackData{
	OnClick: func() {},
}

var ToolbarButton = co.DefineType(&ToolbarButtonComponent{})

type ToolbarButtonComponent struct {
	BaseButtonComponent

	Scope      co.Scope      `co:"scope"`
	Properties co.Properties `co:"properties"`

	icon       *ui.Image
	text       string
	isEnabled  bool
	isSelected bool
}

func (c *ToolbarButtonComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties, toolbarButtonDefaultData)
	c.icon = data.Icon
	c.text = data.Text
	c.isEnabled = !data.Enabled.Specified || data.Enabled.Value
	c.isSelected = data.Selected

	callbackData := co.GetOptionalCallbackData(c.Properties, toolbarButtonDefaultCallbackData)
	c.SetOnClickListener(callbackData.OnClick)
}

func (c *ToolbarButtonComponent) Render() co.Instance {
	// force specific height
	layoutData := co.GetOptionalLayoutData(c.Properties, layout.Data{})
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
					Font:      co.OpenFont(c.Scope, ToolbarButtonFontFile),
					FontSize:  opt.V(float32(ToolbarButtonFontSize)),
					FontColor: opt.V(foregroundColor),
					Text:      c.text,
				})
				co.WithLayoutData(layout.Data{})
			}))
		}
	})
}

func (e *ToolbarButtonComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch e.State() {
	case ButtonStateOver:
		backgroundColor = HoverOverlayColor
	case ButtonStateDown:
		backgroundColor = PressOverlayColor
	default:
		backgroundColor = ui.Transparent()
	}

	bounds := element.Bounds()
	size := sprec.NewVec2(
		float32(bounds.Width),
		float32(bounds.Height),
	)

	if !backgroundColor.Transparent() {
		canvas.Reset()
		canvas.Rectangle(
			sprec.ZeroVec2(),
			size,
		)
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
	if e.isSelected {
		canvas.Reset()
		canvas.Rectangle(
			sprec.NewVec2(0.0, size.Y-ToolbarBottonSelectionHeight),
			sprec.NewVec2(size.X, ToolbarBottonSelectionHeight),
		)
		canvas.Fill(ui.Fill{
			Color: SecondaryColor,
		})
	}
}
