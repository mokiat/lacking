package std

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	DropdownItemIndicatorSize    = 20
	DropdownItemIndicatorPadding = 5
)

type dropdownItemData struct {
	Selected bool
}

type dropdownItemCallbackData struct {
	OnSelected OnActionFunc
}

var dropdownItem = co.Define(&dropdownItemComponent{})

type dropdownItemComponent struct {
	co.BaseComponent
	BaseButtonComponent

	isSelected bool
}

func (c *dropdownItemComponent) OnUpsert() {
	data := co.GetData[dropdownItemData](c.Properties())
	c.isSelected = data.Selected

	callbackData := co.GetCallbackData[dropdownItemCallbackData](c.Properties())
	c.SetOnClickFunc(callbackData.OnSelected)
}

func (c *dropdownItemComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Padding: ui.Spacing{
				Left: DropdownItemIndicatorSize + DropdownItemIndicatorPadding,
			},
			Essence: c,
			Layout:  layout.Fill(),
		})
		co.WithChildren(c.Properties().Children())
	})
}

func (c *dropdownItemComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
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

	if c.isSelected {
		canvas.Reset()
		canvas.Rectangle(
			sprec.ZeroVec2(),
			sprec.NewVec2(float32(DropdownItemIndicatorSize), drawBounds.Height()),
		)
		canvas.Fill(ui.Fill{
			Color: SecondaryColor,
		})
	}

	if !backgroundColor.Transparent() {
		spacing := float32(DropdownItemIndicatorSize + DropdownItemIndicatorPadding)
		canvas.Reset()
		canvas.Rectangle(
			sprec.NewVec2(spacing, 0.0),
			sprec.NewVec2(drawBounds.Width()-spacing, drawBounds.Height()),
		)
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
}
