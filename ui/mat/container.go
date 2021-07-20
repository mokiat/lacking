package mat

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
	t "github.com/mokiat/lacking/ui/template"
)

type ContainerData struct {
	BackgroundColor optional.Color
	Layout          ui.Layout
}

var Container = t.ShallowCached(t.Define(func(props t.Properties) t.Instance {
	var (
		data    ContainerData
		essence *containerEssence
	)
	props.InjectData(&data)

	t.UseState(func() interface{} {
		return &containerEssence{}
	}).Inject(&essence)

	if data.BackgroundColor.Specified {
		essence.backgroundColor = data.BackgroundColor.Value
	} else {
		essence.backgroundColor = ui.Transparent()
	}

	return t.New(Element, func() {
		t.WithData(ElementData{
			Essence: essence,
			Layout:  data.Layout,
		})
		t.WithLayoutData(props.LayoutData())
		t.WithChildren(props.Children())
	})
}))

var _ ui.ElementRenderHandler = (*containerEssence)(nil)

type containerEssence struct {
	backgroundColor ui.Color
}

func (l *containerEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	if !l.backgroundColor.Transparent() {
		canvas.SetSolidColor(l.backgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	}
}
