package component

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/layout"
)

// ElementData is the struct that should be used when configuring
// an Element component's data.
type ElementData struct {

	// Essence is a mechanism to register hooks for various Element events.
	Essence ui.Essence

	// Bounds, if specified, forces the Element to have the specified bounds.
	// This should only be used when a layout system is not in place.
	Bounds opt.T[ui.Bounds]

	// IdealSize is a hint to the layout system about the ideal size of
	// the Element.
	IdealSize opt.T[ui.Size]

	// Padding specifies the spacing between the Element's border and its
	// content.
	Padding ui.Spacing

	// Layout specifies how children of the Element should be laid out.
	Layout ui.Layout

	// Enabled, if specified, controls whether the Element is enabled
	// or disabled.
	Enabled opt.T[bool]

	// Visible, if specified, controls whether the Element is visible
	// or hidden. A visible element still takes space in the layout.
	Visible opt.T[bool]

	// CanAutoFocus, if specified, controls whether the Element can
	// automatically receive focus when clicked or touched.
	CanAutoFocus opt.T[bool]

	// CanAutoUnfocus, if specified, controls whether the Element can
	// automatically lose focus when another Element is clicked or touched.
	CanAutoUnfocus opt.T[bool]

	// Focused, if specified, controls whether the Element should have focus.
	Focused opt.T[bool]

	// CreateFocused controls whether the Element should request focus when
	// created.
	CreateFocused bool
}

var elementDefaultData = ElementData{
	Layout: layout.Fill(),
}

// Element represents the most basic component, which is translated
// to a UI Element. All higher-order components eventually boil down to an
// Element.
var Element = Define[*elementComponent]()

type elementComponent struct {
	BaseComponent
}

func (c *elementComponent) OnCreate() {
	data := GetOptionalData(c.Properties(), elementDefaultData)
	if data.CreateFocused {
		Window(c.Scope()).GrantFocus(c.Element())
	}
}

func (c *elementComponent) OnUpsert() {
	data := GetOptionalData(c.Properties(), elementDefaultData)
	element := c.Element()
	element.SetName(c.Name())
	element.SetEssence(data.Essence)
	if data.Enabled.Specified {
		element.SetEnabled(data.Enabled.Value)
	}
	if data.Visible.Specified {
		element.SetVisible(data.Visible.Value)
	}
	if data.CanAutoFocus.Specified {
		element.SetCanAutoFocus(data.CanAutoFocus.Value)
	}
	if data.CanAutoUnfocus.Specified {
		element.SetCanAutoUnfocus(data.CanAutoUnfocus.Value)
	}
	if data.Bounds.Specified {
		element.SetBounds(data.Bounds.Value)
	}
	if data.IdealSize.Specified {
		element.SetIdealSize(data.IdealSize.Value)
	}
	element.SetLayout(data.Layout)
	element.SetPadding(data.Padding)
	element.SetLayoutConfig(c.Properties().LayoutData())
	if focused, ok := data.Focused.Unwrap(); ok {
		switch {
		case focused && !element.IsFocused():
			element.Focus()
		case !focused && element.IsFocused():
			element.Window().DiscardFocus()
		}
	}
}

func (c *elementComponent) OnDelete() {
	c.Element().Destroy()
}

func (c *elementComponent) Render() Instance {
	return Instance{}
}
