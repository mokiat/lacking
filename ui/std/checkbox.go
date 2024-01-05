package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

const (
	CheckboxContentSpacing    = 5
	CheckboxIconSize          = 24
	CheckboxFontSize          = float32(20)
	CheckboxFontFile          = "ui:///roboto-regular.ttf"
	CheckboxCheckedIconFile   = "ui:///checked.png"
	CheckboxUncheckedIconFile = "ui:///unchecked.png"
)

var Checkbox = co.Define(&checkboxComponent{})

type CheckboxData struct {
	Checked bool
	Label   string
}

type CheckboxCallbackData struct {
	OnToggle func(bool)
}

type checkboxComponent struct {
	co.BaseComponent
	BaseButtonComponent

	font           *ui.Font
	checkedImage   *ui.Image
	uncheckedImage *ui.Image

	label   string
	checked bool

	onToggle func(bool)
}

func (c *checkboxComponent) OnCreate() {
	c.font = co.OpenFont(c.Scope(), CheckboxFontFile)
	c.checkedImage = co.OpenImage(c.Scope(), CheckboxCheckedIconFile)
	c.uncheckedImage = co.OpenImage(c.Scope(), CheckboxUncheckedIconFile)
	c.SetOnClickFunc(c.handleOnClick)
}

func (c *checkboxComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), CheckboxData{})
	c.checked = data.Checked
	c.label = data.Label

	callbackData := co.GetOptionalCallbackData(c.Properties(), CheckboxCallbackData{})
	c.onToggle = callbackData.OnToggle
	if c.onToggle == nil {
		c.onToggle = func(bool) {}
	}
}

func (c *checkboxComponent) Render() co.Instance {
	return co.New(Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(ElementData{
			Essence: c,
			Layout: layout.Horizontal(layout.HorizontalSettings{
				ContentAlignment: layout.VerticalAlignmentCenter,
				ContentSpacing:   CheckboxContentSpacing,
			}),
		})

		co.WithChild("icon", co.New(Picture, func() {
			co.WithLayoutData(layout.Data{
				Width:  opt.V(CheckboxIconSize),
				Height: opt.V(CheckboxIconSize),
			})
			co.WithData(PictureData{
				Image:      c.icon(),
				ImageColor: opt.V(PrimaryLightColor),
				Mode:       ImageModeFit,
			})
		}))

		co.WithChild("label", co.New(Label, func() {
			// co.WithLayoutData(layout.Data{
			// 	Height: opt.V(32),
			// })
			co.WithData(LabelData{
				Font:      c.font,
				FontSize:  opt.V(CheckboxFontSize),
				FontColor: opt.V(OnSurfaceColor),
				Text:      c.label,
			})
		}))

	})
}

func (c *checkboxComponent) icon() *ui.Image {
	if c.checked {
		return c.checkedImage
	}
	return c.uncheckedImage
}

func (c *checkboxComponent) handleOnClick() {
	c.onToggle(!c.checked)
	c.Invalidate()
}
