package mat

import (
	"github.com/mokiat/lacking/ui/optional"
	t "github.com/mokiat/lacking/ui/template"
)

type ViewData struct {
	BackgroundColor optional.Color
	FontSrc         string
}

var View = t.ShallowCached(t.Define(func(props t.Properties) t.Instance {
	var data ViewData
	props.InjectData(&data)

	if data.FontSrc != "" {
		t.OpenFontCollection(data.FontSrc)
	}

	return t.New(Container, func() {
		t.WithData(ContainerData{
			BackgroundColor: data.BackgroundColor,
			Layout:          NewFillLayout(),
		})
		t.WithLayoutData(props.LayoutData())
		t.WithChildren(props.Children())
	})
}))
