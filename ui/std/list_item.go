package std

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	ListItemPadding = 2
)

// ListItemData holds the data for a ListItem component.
type ListItemData struct {
	Selected bool
}

var listItemDefaultData = ListItemData{}

// ListItemCallbackData holds the callback data for a ListItem component.
type ListItemCallbackData struct {
	OnSelected OnActionFunc
}

var listItemDefaultCallbackData = ListItemCallbackData{
	OnSelected: func() {},
}

// ListItem represents a component to be displayed in a List.
var ListItem = co.Define(&listItemComponent{})

type listItemComponent struct {
	BaseButtonComponent

	Properties co.Properties `co:"properties"`

	isSelected bool
}

func (c *listItemComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties, listItemDefaultData)
	c.isSelected = data.Selected

	callbackData := co.GetOptionalCallbackData(c.Properties, listItemDefaultCallbackData)
	c.SetOnClickFunc(callbackData.OnSelected)
}

func (c *listItemComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties.LayoutData())
		co.WithData(co.ElementData{
			Padding: ui.Spacing{
				Left:   ListItemPadding,
				Right:  ListItemPadding,
				Top:    ListItemPadding,
				Bottom: ListItemPadding,
			},
			Essence: c,
			Layout:  layout.Fill(),
		})
		co.WithChildren(c.Properties.Children())
	})
}

func (c *listItemComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch c.State() {
	case ButtonStateOver:
		backgroundColor = HoverOverlayColor
	case ButtonStateDown:
		backgroundColor = PressOverlayColor
	default:
		backgroundColor = ui.Transparent()
	}
	if c.isSelected {
		backgroundColor = SecondaryColor
	}

	size := element.Bounds().Size
	width := float32(size.Width)
	height := float32(size.Height)

	canvas.Reset()
	canvas.SetStrokeSize(1.0)
	canvas.SetStrokeColor(OutlineColor)
	canvas.Rectangle(
		sprec.ZeroVec2(),
		sprec.NewVec2(width, height),
	)
	if !backgroundColor.Transparent() {
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
	canvas.Stroke()
}
