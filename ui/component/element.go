package component

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/util/optional"
)

// ElementData is the struct that should be used when configuring
// an Element component's data.
type ElementData struct {
	Essence   ui.Essence
	Enabled   optional.V[bool]
	Visible   optional.V[bool]
	Focusable optional.V[bool]
	IdealSize optional.V[ui.Size]
	Padding   ui.Spacing
	Layout    ui.Layout
}

// Element represents the most basic component, which is translated
// to a UI Element. All higher-order components eventually boil down to an
// Element.
var Element = Define(func(props Properties) Instance {
	element := UseState(func() *ui.Element {
		return uiCtx.CreateElement()
	}).Get()

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
	if data.Focusable.Specified {
		element.SetFocusable(data.Focusable.Value)
	}
	if data.IdealSize.Specified {
		element.SetIdealSize(data.IdealSize.Value)
	}
	element.SetLayout(data.Layout)
	element.SetPadding(data.Padding)
	element.SetLayoutConfig(props.LayoutData())

	return Instance{
		element:  element,
		children: props.Children(),
	}
})
