package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	AccordionHeaderPadding        = 2
	AccordionHeaderContentSpacing = 5
	AccordionHeaderIconSize       = 24
	AccordionHeaderFontSize       = float32(20)
	AccordionHeaderFontFile       = "ui:///roboto-regular.ttf"
	AccordionExpandedIconFile     = "ui:///expanded.png"
	AccordionCollapsedIconFile    = "ui:///collapsed.png"
)

// AccordionData holds the data for an Accordion container.
type AccordionData struct {
	Title    string
	Expanded bool
}

var accordionDefaultData = AccordionData{}

// AccordionCallbackData holds the callback data for an Accordion container.
type AccordionCallbackData struct {
	OnToggle func(bool)
}

var accordionDefaultCallbackData = AccordionCallbackData{
	OnToggle: func(bool) {},
}

var Accordion = co.Define(&accordionComponent{})

type accordionComponent struct {
	co.BaseComponent
	BaseButtonComponent

	expandedIcon  *ui.Image
	collapsedIcon *ui.Image

	title      string
	isExpanded bool

	onToggle func(bool)
}

func (c *accordionComponent) OnCreate() {
	c.SetOnClickFunc(c.handleOnClick)
	c.expandedIcon = co.OpenImage(c.Scope(), AccordionExpandedIconFile)
	c.collapsedIcon = co.OpenImage(c.Scope(), AccordionCollapsedIconFile)
}

func (c *accordionComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), accordionDefaultData)
	c.isExpanded = data.Expanded
	c.title = data.Title

	callbackData := co.GetOptionalCallbackData(c.Properties(), accordionDefaultCallbackData)
	c.onToggle = callbackData.OnToggle
	if c.onToggle == nil {
		c.onToggle = func(bool) {}
	}
}

func (c *accordionComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Layout: layout.Vertical(layout.VerticalSettings{
				ContentAlignment: layout.HorizontalAlignmentLeft,
			}),
		})

		co.WithChild("header", co.New(co.Element, func() {
			co.WithLayoutData(layout.Data{
				GrowHorizontally: true,
			})
			co.WithData(co.ElementData{
				Essence: c,
				Padding: ui.Spacing{
					Left:   AccordionHeaderPadding,
					Right:  AccordionHeaderPadding,
					Top:    AccordionHeaderPadding,
					Bottom: AccordionHeaderPadding,
				},
				Layout: layout.Horizontal(layout.HorizontalSettings{
					ContentAlignment: layout.VerticalAlignmentCenter,
					ContentSpacing:   AccordionHeaderContentSpacing,
				}),
			})

			co.WithChild("icon", co.New(Picture, func() {
				co.WithData(PictureData{
					Image:      c.icon(),
					ImageColor: opt.V(OnPrimaryLightColor),
					Mode:       ImageModeFit,
				})
				co.WithLayoutData(layout.Data{
					Width:  opt.V(AccordionHeaderIconSize),
					Height: opt.V(AccordionHeaderIconSize),
				})
			}))

			co.WithChild("title", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(c.Scope(), AccordionHeaderFontFile),
					FontSize:  opt.V(AccordionHeaderFontSize),
					FontColor: opt.V(OnPrimaryLightColor),
					Text:      c.title,
				})
			}))
		}))

		if c.isExpanded {
			for _, child := range c.Properties().Children() {
				co.WithChild(child.Key(), child)
			}
		}
	})
}

func (c *accordionComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch c.State() {
	case ButtonStateOver:
		backgroundColor = PrimaryLightColor.Overlay(HoverOverlayColor)
	case ButtonStateDown:
		backgroundColor = PrimaryLightColor.Overlay(PressOverlayColor)
	default:
		backgroundColor = PrimaryLightColor
	}

	drawBounds := canvas.DrawBounds(element, false)

	canvas.Reset()
	canvas.SetStrokeSize(1.0)
	canvas.SetStrokeColor(OutlineColor)
	canvas.Rectangle(
		drawBounds.Position,
		drawBounds.Size,
	)
	if !backgroundColor.Transparent() {
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
	canvas.Stroke()
}

func (c *accordionComponent) icon() *ui.Image {
	if c.isExpanded {
		return c.expandedIcon
	}
	return c.collapsedIcon
}

func (c *accordionComponent) handleOnClick() {
	c.onToggle(!c.isExpanded)
}
