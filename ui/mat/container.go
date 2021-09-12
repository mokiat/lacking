package mat

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/optional"
)

type ContainerData struct {
	BackgroundColor optional.Color
	Layout          ui.Layout
}

var defaultContainerData = ContainerData{
	Layout: NewFillLayout(),
}

var Container = co.ShallowCached(co.Define(func(props co.Properties) co.Instance {
	var (
		data    ContainerData
		essence *containerEssence
	)
	props.InjectOptionalData(&data, defaultContainerData)

	co.UseState(func() interface{} {
		return &containerEssence{}
	}).Inject(&essence)

	if data.BackgroundColor.Specified {
		essence.backgroundColor = data.BackgroundColor.Value
	} else {
		essence.backgroundColor = ui.Transparent()
	}

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Layout:  data.Layout,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
}))

var _ ui.ElementRenderHandler = (*containerEssence)(nil)

type containerEssence struct {
	backgroundColor ui.Color
}

func (l *containerEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	if !l.backgroundColor.Transparent() {
		canvas.Shape().Begin(ui.Fill{
			Color: l.backgroundColor,
		})
		canvas.Shape().Rectangle(
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
		canvas.Shape().End()
	}
}
