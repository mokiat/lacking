package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	DropdownFontSize = float32(18)
	DropdownFontFile = "ui:///roboto-regular.ttf"
	DropdownIconSize = 24
	DropdownIconFile = "ui:///expanded.png"
)

// DropdownData holds the data for a Dropdown container.
type DropdownData struct {
	Items       []DropdownItem
	SelectedKey any
}

var dropdownDefaultData = DropdownData{}

// DropdownItem represents a single dropdown entry.
type DropdownItem struct {
	Key   any
	Label string
}

// DropdownCallbackData holds the callback data for a Dropdown container.
type DropdownCallbackData struct {
	OnItemSelected func(key any)
}

var dropdownDefaultCallbackData = DropdownCallbackData{
	OnItemSelected: func(key any) {},
}

var Dropdown = co.Define(&DropdownComponent{})

type DropdownComponent struct {
	BaseButtonComponent

	Scope      co.Scope      `co:"scope"`
	Properties co.Properties `co:"properties"`
	Invalidate func()        `co:"invalidate"`

	items           []DropdownItem
	selectedItemKey any

	overlay co.Overlay

	onItemSelected func(key any)
}

func (c *DropdownComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties, dropdownDefaultData)
	c.items = data.Items
	c.selectedItemKey = data.SelectedKey

	callbackData := co.GetOptionalCallbackData(c.Properties, dropdownDefaultCallbackData)
	c.onItemSelected = callbackData.OnItemSelected

	c.SetOnClickFunc(c.onOpen)
}

func (c *DropdownComponent) Render() co.Instance {
	label := ""
	for _, item := range c.items {
		if item.Key == c.selectedItemKey {
			label = item.Label
		}
	}

	return co.New(co.Element, func() {
		co.WithData(co.ElementData{
			Essence: c,
			Layout:  layout.Anchor(),
		})
		co.WithLayoutData(c.Properties.LayoutData())

		co.WithChild("label", co.New(Label, func() {
			co.WithLayoutData(layout.Data{
				Left:           opt.V(0),
				Right:          opt.V(DropdownIconSize),
				VerticalCenter: opt.V(0),
			})
			co.WithData(LabelData{
				Font:      co.OpenFont(c.Scope, DropdownFontFile),
				FontSize:  opt.V(DropdownFontSize),
				FontColor: opt.V(OnSurfaceColor),
				Text:      label,
			})
		}))

		co.WithChild("button", co.New(Picture, func() {
			co.WithLayoutData(layout.Data{
				Width:          opt.V(DropdownIconSize),
				Height:         opt.V(DropdownIconSize),
				Right:          opt.V(0),
				VerticalCenter: opt.V(0),
			})
			co.WithData(PictureData{
				Image:      co.OpenImage(c.Scope, DropdownIconFile),
				ImageColor: opt.V(OnSurfaceColor),
				Mode:       ImageModeFit,
			})
		}))
	})
}

func (c *DropdownComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch c.State() {
	case ButtonStateOver:
		backgroundColor = HoverOverlayColor
	case ButtonStateDown:
		backgroundColor = PressOverlayColor
	default:
		backgroundColor = ui.Transparent()
	}

	size := element.Bounds().Size
	width := float32(size.Width)
	height := float32(size.Height)

	canvas.Reset()
	canvas.SetStrokeSize(2.0)
	canvas.SetStrokeColor(PrimaryLightColor)
	canvas.RoundRectangle(
		sprec.ZeroVec2(),
		sprec.NewVec2(width, height),
		sprec.NewVec4(5.0, 5.0, 5.0, 5.0),
	)
	if !backgroundColor.Transparent() {
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
	canvas.Stroke()
}

func (c *DropdownComponent) onSelected(key any) {
	c.onClose()
	c.onItemSelected(key)
	c.Invalidate()
}

func (c *DropdownComponent) onOpen() {
	c.overlay = co.OpenOverlay(c.Scope, co.New(dropdownList, func() {
		co.WithData(c.Properties.Data())
		co.WithCallbackData(dropdownListCallbackData{
			OnSelected: c.onSelected,
			OnClose:    c.onClose,
		})
	}))
	c.Invalidate()
}

func (c *DropdownComponent) onClose() {
	c.overlay.Close()
	c.overlay = nil
	c.Invalidate()
}
