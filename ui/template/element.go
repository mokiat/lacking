package template

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
)

type ElementData struct {
	Essence      ui.Essence
	Enabled      optional.Bool
	Visible      optional.Bool
	Materialized optional.Bool
}

var Element = Plain(func(props Properties) Instance {
	var element *ui.Element

	Once(func() {
		element = uiCtx.CreateElement()
	})

	Defer(func() {
		element.Destroy()
	})

	var data ElementData
	props.InjectData(&data)

	element.SetEssence(data.Essence)
	if data.Enabled.Specified {
		element.SetEnabled(data.Enabled.Value)
	}
	if data.Visible.Specified {
		element.SetVisible(data.Visible.Value)
	}
	if data.Materialized.Specified {
		element.SetMaterialized(data.Materialized.Value)
	}
	element.SetLayoutConfig(props.LayoutData())

	return Instance{
		element:  element,
		children: props.Children(),
	}
})
