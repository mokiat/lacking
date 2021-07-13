package mat

import (
	"github.com/mokiat/lacking/ui"
	t "github.com/mokiat/lacking/ui/template"
)

type ContainerData struct {
	BackgroundColor ui.Color
	Layout          Layout
}

var Container = t.NewComponentType(namespace, "Container", func(ctx *ui.Context) t.Component {
	return t.FunctionalComponent(func(rc t.RenderContext) t.Instance {
		return rc.Instance(Element, rc.Key(), func() {
			rc.WithData(t.ElementData{
				Essence: &containerEssence{
					ContainerData: rc.Data().(ContainerData),
				},
			})
			rc.WithLayoutData(rc.LayoutData())
			rc.WithChildren(rc.Children())
		})
	})
})

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
