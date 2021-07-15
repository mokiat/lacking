package mat

import (
	"github.com/mokiat/lacking/ui"
	t "github.com/mokiat/lacking/ui/template"
)

type ContainerData struct {
	BackgroundColor ui.Color
	Layout          Layout
}

var Container = t.ShallowCached(t.Plain(func(props t.Properties) t.Instance {
	var data ContainerData
	props.InjectData(&data)

	return t.New(Element, func() {
		t.WithData(ElementData{
			Essence: &containerEssence{
				ContainerData: data,
			},
		})
		t.WithLayoutData(props.LayoutData())
		t.WithChildren(props.Children())
	})
}))

var _ ui.ElementResizeHandler = (*containerEssence)(nil)
var _ ui.ElementRenderHandler = (*containerEssence)(nil)

type containerEssence struct {
	ContainerData
}

func (l *containerEssence) OnResize(element *ui.Element, bounds ui.Bounds) {
	l.Layout.Apply(element)
}

func (l *containerEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	if !l.BackgroundColor.Transparent() {
		canvas.SetSolidColor(l.BackgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	}
}
