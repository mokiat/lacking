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
	layoutData.Height = optional.NewInt(ToolbarHeight)

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
	size := element.Bounds().Size

	canvas.Shape().Begin(ui.Fill{
		Color: ToolbarColor,
	})
	canvas.Shape().Rectangle(
		ui.NewPosition(0, 0),
		size,
	)
	canvas.Shape().End()

	stroke := ui.Stroke{
		Color: ToolbarBorderColor,
		Size:  2,
	}
	canvas.Contour().Begin()
	canvas.Contour().MoveTo(ui.NewPosition(0, size.Height), stroke)
	canvas.Contour().LineTo(ui.NewPosition(size.Width, size.Height), stroke)
	canvas.Contour().End()
}
