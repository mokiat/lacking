package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

type ContainerData struct {
	BackgroundColor opt.T[ui.Color]
	Padding         ui.Spacing
	Layout          ui.Layout
}

var defaultContainerData = ContainerData{
	Layout: NewFillLayout(),
}

var Container = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data ContainerData
	)
	props.InjectOptionalData(&data, defaultContainerData)

	essence := co.UseState(func() *containerEssence {
		return &containerEssence{}
	}).Get()

	if data.BackgroundColor.Specified {
		essence.backgroundColor = data.BackgroundColor.Value
	} else {
		essence.backgroundColor = ui.Transparent()
	}

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Padding: data.Padding,
			Layout:  data.Layout,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.ElementRenderHandler = (*containerEssence)(nil)

type containerEssence struct {
	backgroundColor ui.Color
}

func (l *containerEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	if !l.backgroundColor.Transparent() {
		canvas.Reset()
		canvas.Rectangle(
			sprec.NewVec2(0, 0),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Fill(ui.Fill{
			Color: l.backgroundColor,
		})
	}
}
