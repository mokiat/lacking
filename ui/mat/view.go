package mat

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
	t "github.com/mokiat/lacking/ui/template"
)

type ViewData struct {
	BackgroundColor optional.Color
	FontSrc         string
}

var View = t.ShallowCached(t.Plain(func(props t.Properties) t.Instance {
	var (
		data    ViewData
		essence *viewEssence
	)
	props.InjectData(&data)

	t.UseState(func() interface{} {
		return &viewEssence{
			backgroundColor: ui.Transparent(),
		}
	}).Inject(&essence)

	if data.BackgroundColor.Specified {
		essence.backgroundColor = data.BackgroundColor.Value
	} else {
		essence.backgroundColor = ui.Transparent()
	}
	if data.FontSrc != "" {
		t.OpenFontCollection(data.FontSrc)
	}

	return t.New(Element, func() {
		t.WithData(ElementData{
			Essence: essence,
		})
		t.WithLayoutData(props.LayoutData())
		t.WithChildren(props.Children())
	})
}))

var _ ui.ElementResizeHandler = (*viewEssence)(nil)
var _ ui.ElementRenderHandler = (*viewEssence)(nil)

type viewEssence struct {
	backgroundColor ui.Color
}

func (v *viewEssence) OnResize(element *ui.Element, bounds ui.Bounds) {
	contentBounds := element.ContentBounds()
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		childElement.SetBounds(contentBounds)
	}
}

func (v *viewEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	if !v.backgroundColor.Transparent() {
		canvas.SetSolidColor(v.backgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	}
}
