package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

var (
	DropdownItemIndicatorSize    = 20
	DropdownItemIndicatorPadding = 5
)

type dropdownItemData struct {
	Selected bool
}

type dropdownItemCallbackData struct {
	OnSelected func()
}

var dropdownItem = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data         = co.GetData[dropdownItemData](props)
		callbackData = co.GetCallbackData[dropdownItemCallbackData](props)
	)

	essence := co.UseState(func() *dropdownItemEssence {
		return &dropdownItemEssence{
			ButtonBaseEssence: NewButtonBaseEssence(callbackData.OnSelected),
		}
	}).Get()
	essence.selected = data.Selected

	return co.New(Element, func() {
		co.WithData(ElementData{
			Padding: ui.Spacing{
				Left: DropdownItemIndicatorSize + DropdownItemIndicatorPadding,
			},
			Essence: essence,
			Layout:  NewFillLayout(),
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.ElementRenderHandler = (*dropdownItemEssence)(nil)

type dropdownItemEssence struct {
	*ButtonBaseEssence

	selected bool
}

func (e *dropdownItemEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch e.State() {
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

	if e.selected {
		canvas.Reset()
		canvas.Rectangle(
			sprec.ZeroVec2(),
			sprec.NewVec2(float32(DropdownItemIndicatorSize), height),
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
			sprec.NewVec2(width-spacing, height),
		)
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
}
