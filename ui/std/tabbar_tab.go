package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
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

var tabbarTabDefaultData = TabbarTabData{}

// TabbarTabCallbackData holds the callback data for a TabbarTab component.
type TabbarTabCallbackData struct {
	OnClick OnClickFunc
	OnClose OnClickFunc
}

var tabbarTabDefaultCallbackData = TabbarTabCallbackData{
	OnClick: func() {},
	OnClose: func() {},
}

var TabbarTab = co.DefineType(&TabbarTabComponent{})

type TabbarTabComponent struct {
	Scope      co.Scope      `co:"scope"`
	Properties co.Properties `co:"properties"`

	main  tabbarTabMainComponent
	close tabbarTabCloseComponent

	icon *ui.Image
	text string
}

func (c *TabbarTabComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties, tabbarTabDefaultData)
	c.icon = data.Icon
	c.text = data.Text
	c.main.isSelected = data.Selected

	callbackData := co.GetOptionalCallbackData(c.Properties, tabbarTabDefaultCallbackData)
	c.main.SetOnClickListener(callbackData.OnClick)
	c.close.SetOnClickListener(callbackData.OnClose)
}

func (c *TabbarTabComponent) Render() co.Instance {
	// force specific height
	layoutData := co.GetOptionalLayoutData(c.Properties, layout.Data{})
	layoutData.Height = opt.V(TabbarTabHeight)

	return co.New(Element, func() {
		co.WithLayoutData(layoutData)
		co.WithData(ElementData{
			Essence: &c.main,
			Layout: layout.Horizontal(layout.HorizontalSettings{
				ContentAlignment: layout.VerticalAlignmentCenter,
				ContentSpacing:   TabbarTabContentSpacing,
			}),
			Padding: ui.Spacing{
				Top:   TabbarTabTopPadding,
				Left:  TabbarTabSidePadding,
				Right: TabbarTabSidePadding,
			},
		})

		if c.icon != nil {
			co.WithChild("icon", co.New(Picture, func() {
				co.WithData(PictureData{
					Image:      c.icon,
					ImageColor: opt.V(OnSurfaceColor),
					Mode:       ImageModeFit,
				})
				co.WithLayoutData(layout.Data{
					Width:  opt.V(TabbarTabIconSize),
					Height: opt.V(TabbarTabIconSize),
				})
			}))
		}

		if c.text != "" {
			co.WithChild("text", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(c.Scope, TabbarTabFontFile),
					FontSize:  opt.V(TabbarTabFontSize),
					FontColor: opt.V(OnSurfaceColor),
					Text:      c.text,
				})
			}))
		}

		if c.main.isSelected {
			co.WithChild("close", co.New(Element, func() {
				co.WithData(ElementData{
					Essence: &c.close,
					Layout:  layout.Fill(),
				})

				co.WithLayoutData(layout.Data{
					Width:  opt.V(TabbarTabIconSize),
					Height: opt.V(TabbarTabIconSize),
				})

				co.WithChild("icon", co.New(Picture, func() {
					co.WithData(PictureData{
						Image:      co.OpenImage(c.Scope, TabbarTabCloseIconFile),
						ImageColor: opt.V(OnSurfaceColor),
						Mode:       ImageModeFit,
					})
				}))
			}))
		}
	})
}

type tabbarTabMainComponent struct {
	BaseButtonComponent
	isSelected bool
}

func (c tabbarTabMainComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	if c.isSelected {
		backgroundColor = SurfaceColor
	} else {
		switch c.State() {
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

type tabbarTabCloseComponent struct {
	BaseButtonComponent
}

func (c tabbarTabCloseComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch c.State() {
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
