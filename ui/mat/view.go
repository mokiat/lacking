package mat

import (
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/optional"
)

type ViewData struct {
	BackgroundColor optional.Color
	FontSrc         string
}

var View = co.ShallowCached(co.Define(func(props co.Properties) co.Instance {
	var data ViewData
	props.InjectData(&data)

	if data.FontSrc != "" {
		co.OpenFontCollection(data.FontSrc)
	}

	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: data.BackgroundColor,
			Layout:          NewFillLayout(),
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
}))
