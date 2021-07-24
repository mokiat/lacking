package component

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
)

// ElementData is the struct that should be used when configuring
// an Element component's data.
type ElementData struct {
	Essence      ui.Essence
	Enabled      optional.Bool
	Visible      optional.Bool
	Materialized optional.Bool
	Padding      ui.Spacing
	Layout       ui.Layout
}

// Element represents the most basic component, which is translated
// to an ui Element. All higher-order components will eventually
// boil down to an Element.
var Element = Define(func(props Properties) Instance {
	var element *ui.Element

	UseState(func() interface{} {
		return uiCtx.CreateElement()
	}).Inject(&element)

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
	element.SetLayout(data.Layout)
	element.SetPadding(data.Padding)
	element.SetLayoutConfig(props.LayoutData())

	return Instance{
		element:  element,
		children: props.Children(),
	}
})
