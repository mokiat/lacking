package component

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
)

// ElementData is the struct that should be used when configuring
// an Element component's data.
type ElementData struct {
	Essence   ui.Essence
	Enabled   opt.T[bool]
	Visible   opt.T[bool]
	Focusable opt.T[bool]
	Focused   opt.T[bool]
	Bounds    opt.T[ui.Bounds]
	IdealSize opt.T[ui.Size]
	Padding   ui.Spacing
	Layout    ui.Layout
}

// Element represents the most basic component, which is translated
// to a UI Element. All higher-order components eventually boil down to an
// Element.
var Element = Define(func(props Properties, scope Scope) Instance {
	data := GetData[ElementData](props)

	element := UseState(func() *ui.Element {
		return Window(scope).CreateElement()
	}).Get()

	Once(func() {
		if data.Focused.Specified && data.Focused.Value {
			Window(scope).GrantFocus(element)
		}
	})

	Defer(func() {
		element.Destroy()
	})

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
	if data.Bounds.Specified {
		element.SetBounds(data.Bounds.Value)
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
