package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	ToolbarLogoSidePadding     = 4
	ToolbarLogoContentSpacing  = 5
	ToolbarLogoIconSize        = 24
	ToolbarLogoFontFile        = "ui:///roboto-bold.ttf"
	ToolbarLogoFontSize        = float32(20)
	ToolbarLogoSelectionHeight = float32(5.0)
)

// ToolbarLogoData holds the data for a ToolbarLogo component.
type ToolbarLogoData struct {
	Image *ui.Image
	Text  string
}

var toolbarLogoDefaultData = ToolbarLogoData{
	Text: "N/A",
}

// ToolbarLogo is used to display a logo on the Toolbar. Usually this would be
// the first item in a Toolbar.
var ToolbarLogo = co.Define(&toolbarLogoComponent{})

type toolbarLogoComponent struct {
	co.BaseComponent

	image *ui.Image
	text  string
}

func (c *toolbarLogoComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), toolbarLogoDefaultData)
	c.image = data.Image
	c.text = data.Text
}

func (c *toolbarLogoComponent) Render() co.Instance {
	// force specific height
	layoutData := co.GetOptionalLayoutData(c.Properties(), layout.Data{})
	layoutData.Height = opt.V(ToolbarItemHeight)

	return co.New(co.Element, func() {
		co.WithLayoutData(layoutData)
		co.WithData(co.ElementData{
			Essence: c,
			Layout: layout.Horizontal(layout.HorizontalSettings{
				ContentAlignment: layout.VerticalAlignmentCenter,
				ContentSpacing:   ToolbarLogoContentSpacing,
			}),
			Padding: ui.Spacing{
				Left:  ToolbarLogoSidePadding,
				Right: ToolbarLogoSidePadding,
			},
		})

		if c.image != nil {
			co.WithChild("image", co.New(Picture, func() {
				co.WithData(PictureData{
					Image:      c.image,
					ImageColor: opt.V(ui.White()),
					Mode:       ImageModeFit,
				})
				co.WithLayoutData(layout.Data{
					Width:  opt.V(ToolbarLogoIconSize),
					Height: opt.V(ToolbarLogoIconSize),
				})
			}))
		}

		if c.text != "" {
			co.WithChild("text", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(c.Scope(), ToolbarLogoFontFile),
					FontSize:  opt.V(float32(ToolbarLogoFontSize)),
					FontColor: opt.V(OnSurfaceColor),
					Text:      c.text,
				})
				co.WithLayoutData(layout.Data{})
			}))
		}
	})
}
