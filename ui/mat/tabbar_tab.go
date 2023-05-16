package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/mat/layout"
)

var (
	TabbarTabHeight         = TabbarHeight
	TabbarTabTopPadding     = 5
	TabbarTabSidePadding    = 10
	TabbarTabContentSpacing = 5
	TabbarTabIconSize       = 24
	TabbarTabFontFile       = "mat:///roboto-regular.ttf"
	TabbarTabFontSize       = float32(20)
	TabbarTabCloseIconFile  = "mat:///close.png"
	TabbarTabRadius         = float32(15)
)

// TabbarTabData holds the data for a TabbarTab component.
type TabbarTabData struct {
	Icon     *ui.Image
	Text     string
	Selected bool
}

var defaultTabbarTabData = TabbarTabData{}

// TabbarTabCallbackData holds the callback data for a TabbarTab component.
type TabbarTabCallbackData struct {
	OnClick func()
	OnClose func()
}

var defaultTabbarTabCallbackData = TabbarTabCallbackData{
	OnClick: func() {},
	OnClose: func() {},
}

// TabbarTab is a tab component to be placed inside a Tabbar.
var TabbarTab = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data         = co.GetOptionalData(props, defaultTabbarTabData)
		callbackData = co.GetOptionalCallbackData(props, defaultTabbarTabCallbackData)
		layoutData   = co.GetOptionalLayoutData(props, LayoutData{})
	)

	essence := co.UseState(func() *tabbarTabEssence {
		return &tabbarTabEssence{
			ButtonBaseEssence: NewButtonBaseEssence(callbackData.OnClick),
		}
	}).Get()
	essence.selected = data.Selected

	closeButtonEssence := co.UseState(func() *tabbarTabCloseButtonEssence {
		return &tabbarTabCloseButtonEssence{
			ButtonBaseEssence: NewButtonBaseEssence(callbackData.OnClose),
		}
	}).Get()

	// force specific height
	layoutData.Height = opt.V(TabbarTabHeight)

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Layout: NewHorizontalLayout(HorizontalLayoutSettings{
				ContentAlignment: AlignmentCenter,
				ContentSpacing:   TabbarTabContentSpacing,
			}),
			Padding: ui.Spacing{
				Top:   TabbarTabTopPadding,
				Left:  TabbarTabSidePadding,
				Right: TabbarTabSidePadding,
			},
		})
		co.WithLayoutData(layoutData)

		if data.Icon != nil {
			co.WithChild("icon", co.New(Picture, func() {
				co.WithData(PictureData{
					Image:      data.Icon,
					ImageColor: opt.V(OnSurfaceColor),
					Mode:       ImageModeFit,
				})
				co.WithLayoutData(LayoutData{
					Width:  opt.V(TabbarTabIconSize),
					Height: opt.V(TabbarTabIconSize),
				})
			}))
		}

		if data.Text != "" {
			co.WithChild("text", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(scope, TabbarTabFontFile),
					FontSize:  opt.V(TabbarTabFontSize),
					FontColor: opt.V(OnSurfaceColor),
					Text:      data.Text,
				})
			}))
		}

		if data.Selected {
			co.WithChild("close", co.New(Element, func() {
				co.WithData(ElementData{
					Essence: closeButtonEssence,
					Layout:  layout.Fill(),
				})

				co.WithLayoutData(LayoutData{
					Width:  opt.V(TabbarTabIconSize),
					Height: opt.V(TabbarTabIconSize),
				})

				co.WithChild("icon", co.New(Picture, func() {
					co.WithData(PictureData{
						Image:      co.OpenImage(scope, TabbarTabCloseIconFile),
						ImageColor: opt.V(OnSurfaceColor),
						Mode:       ImageModeFit,
					})
				}))
			}))
		}
	})
})

var _ ui.ElementRenderHandler = (*tabbarTabEssence)(nil)

type tabbarTabEssence struct {
	*ButtonBaseEssence
	selected bool
}

func (e *tabbarTabEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	if e.selected {
		backgroundColor = SurfaceColor
	} else {
		switch e.State() {
		case ButtonStateOver:
			backgroundColor = HoverOverlayColor
		case ButtonStateDown:
			backgroundColor = PressOverlayColor
		default:
			backgroundColor = ui.Transparent()
		}
	}

	size := element.Bounds().Size
	width := float32(size.Width)
	height := float32(size.Height)

	if !backgroundColor.Transparent() {
		canvas.Reset()
		canvas.MoveTo(
			sprec.NewVec2(0, height),
		)
		canvas.LineTo(
			sprec.NewVec2(width, height),
		)
		canvas.LineTo(
			sprec.NewVec2(width, TabbarTabRadius),
		)
		canvas.QuadTo(
			sprec.NewVec2(width, 0),
			sprec.NewVec2(width-TabbarTabRadius, 0),
		)
		canvas.LineTo(
			sprec.NewVec2(TabbarTabRadius, 0),
		)
		canvas.QuadTo(
			sprec.NewVec2(0, 0),
			sprec.NewVec2(0, TabbarTabRadius),
		)
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
}

var _ ui.ElementRenderHandler = (*tabbarTabCloseButtonEssence)(nil)

type tabbarTabCloseButtonEssence struct {
	*ButtonBaseEssence
}

func (e *tabbarTabCloseButtonEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch e.State() {
	case ButtonStateOver:
		backgroundColor = HoverOverlayColor
	case ButtonStateDown:
		backgroundColor = PressOverlayColor
	default:
		backgroundColor = ui.Transparent()
	}

	if !backgroundColor.Transparent() {
		size := element.Bounds().Size
		width := float32(size.Width)
		height := float32(size.Height)
		canvas.Reset()
		canvas.Rectangle(
			sprec.ZeroVec2(),
			sprec.NewVec2(width, height),
		)
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
}
