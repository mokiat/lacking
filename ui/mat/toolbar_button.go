package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
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

var defaultToolbarButtonData = ToolbarButtonData{}

// ToolbarButtonCallbackData holds the callback handlers for a
// ToolbarButton component.
type ToolbarButtonCallbackData struct {
	OnClick ClickListener
}

var defaultToolbarButtonCallbackData = ToolbarButtonCallbackData{
	OnClick: func() {},
}

// ToolbarButton is a button component intended to be placed inside a
// Toolbar container.
var ToolbarButton = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data         = co.GetOptionalData(props, defaultToolbarButtonData)
		layoutData   = co.GetOptionalLayoutData(props, LayoutData{})
		callbackData = co.GetOptionalCallbackData(props, defaultToolbarButtonCallbackData)
	)

	essence := co.UseState(func() *toolbarButtonBackgroundEssence {
		return &toolbarButtonBackgroundEssence{
			ButtonBaseEssence: NewButtonBaseEssence(callbackData.OnClick),
		}
	}).Get()
	essence.selected = data.Selected

	// force specific height
	layoutData.Height = opt.V(ToolbarItemHeight)

	foregroundColor := OnSurfaceColor
	if data.Enabled.Specified && !data.Enabled.Value {
		foregroundColor = OutlineColor
	}

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Layout: NewHorizontalLayout(HorizontalLayoutSettings{
				ContentAlignment: AlignmentCenter,
				ContentSpacing:   ToolbarButtonContentSpacing,
			}),
			Padding: ui.Spacing{
				Left:  ToolbarButtonSidePadding,
				Right: ToolbarButtonSidePadding,
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
				co.WithLayoutData(LayoutData{
					Width:  opt.V(ToolbarButtonIconSize),
					Height: opt.V(ToolbarButtonIconSize),
				})
			}))
		}

		if data.Text != "" {
			co.WithChild("text", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(scope, ToolbarButtonFontFile),
					FontSize:  opt.V(float32(ToolbarButtonFontSize)),
					FontColor: opt.V(foregroundColor),
					Text:      data.Text,
				})
				co.WithLayoutData(LayoutData{})
			}))
		}
	})
})

var _ ui.ElementRenderHandler = (*toolbarButtonBackgroundEssence)(nil)

type toolbarButtonBackgroundEssence struct {
	*ButtonBaseEssence
	selected bool
}

func (e *toolbarButtonBackgroundEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
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
	if e.selected {
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
