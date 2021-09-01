package mat

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/optional"
)

var Toolbar = co.ShallowCached(co.Define(func(props co.Properties) co.Instance {
	var (
		essence    *toolbarEssence
		layoutData LayoutData
	)

	co.UseState(func() interface{} {
		return &toolbarEssence{}
	}).Inject(&essence)

	props.InjectOptionalLayoutData(&layoutData, LayoutData{})
	layoutData.Height = optional.NewInt(50)

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
		})
		co.WithLayoutData(layoutData)
		co.WithChildren(props.Children())
	})
}))

var _ ui.ElementRenderHandler = (*toolbarEssence)(nil)

type toolbarEssence struct {
	// TODO: Compose default mouse handling instead of custom impl
	// TODO: Compose default div-like drawing instead of custom impl
}

func (e *toolbarEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	left := 0
	right := element.ContentBounds().Width - 1
	top := 0
	bottom := element.ContentBounds().Height - 1

	stroke := ui.Stroke{
		Color: ui.Red(),
		Size:  4,
	}
	canvas.BeginShape(ui.Fill{
		Rule:            ui.FillRuleSimple,
		BackgroundColor: ui.Teal(),
	})
	canvas.MoveTo(
		ui.NewPosition(left, top),
	)
	canvas.LineTo(
		ui.NewPosition(left, (top+bottom)/2),
		stroke, stroke,
	)
	canvas.QuadTo(
		ui.NewPosition(left, bottom),
		ui.NewPosition((left+right)/2, bottom),
		stroke, stroke,
	)
	// canvas.QuadTo(
	// 	ui.NewPosition(right, bottom),
	// 	ui.NewPosition(right, (top+bottom)/2),
	// 	stroke, stroke,
	// )
	canvas.CubeTo(
		ui.NewPosition(right-(right-left)*1/3, bottom),
		ui.NewPosition(right, bottom-(bottom-top)*1/4),
		ui.NewPosition(right, (top+bottom)/2),
		stroke, stroke,
	)
	// canvas.LineTo(
	// 	ui.NewPosition(right, bottom),
	// 	stroke,
	// 	stroke,
	// )
	canvas.LineTo(
		ui.NewPosition(right, top),
		stroke, stroke,
	)
	canvas.CloseLoop(stroke, stroke)
	canvas.EndShape()

	// canvas.SetStrokeSize(4)
	// canvas.SetStrokeColor(ui.Red())
	// canvas.SetSolidColor(ui.Teal())
	// canvas.SetBackgroundImage(...) // TODO
	// canvas.SetShadow() // TODO
	// canvas.LineTo((left+right)/2, bottom)

	// canvas.SetSolidColor(ui.RGB(0, 25, 128))
	// canvas.FillRectangle(ui.NewPosition(0, 0), element.ContentBounds().Size)
}
