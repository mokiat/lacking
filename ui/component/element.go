package component

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/layout"
)

// ElementData is the struct that should be used when configuring
// an Element component's data.
type ElementData struct {
	Reference **ui.Element
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

var elementDefaultData = ElementData{
	Layout: layout.Fill(),
}

// Element represents the most basic component, which is translated
// to a UI Element. All higher-order components eventually boil down to an
// Element.
var Element = Define(&ElementComponent{})

type ElementComponent struct {
	Scope      Scope      `co:"scope"`
	Properties Properties `co:"properties"`

	element *ui.Element
}

func (c *ElementComponent) OnCreate() {
	c.element = Window(c.Scope).CreateElement()

	c.OnUpdate()

	data := GetOptionalData(c.Properties, elementDefaultData)
	if data.Focused.Specified && data.Focused.Value {
		Window(c.Scope).GrantFocus(c.element)
	}
}

func (c *ElementComponent) OnUpdate() {
	data := GetOptionalData(c.Properties, elementDefaultData)
	if data.Reference != nil {
		*data.Reference = c.element
	}

	c.element.SetEssence(data.Essence)
	if data.Enabled.Specified {
		c.element.SetEnabled(data.Enabled.Value)
	}
	if data.Visible.Specified {
		c.element.SetVisible(data.Visible.Value)
	}
	if data.Focusable.Specified {
		c.element.SetFocusable(data.Focusable.Value)
	}
	if data.Bounds.Specified {
		c.element.SetBounds(data.Bounds.Value)
	}
	if data.IdealSize.Specified {
		c.element.SetIdealSize(data.IdealSize.Value)
	}
	c.element.SetLayout(data.Layout)
	c.element.SetPadding(data.Padding)
	c.element.SetLayoutConfig(c.Properties.LayoutData())
}

func (c *ElementComponent) OnDelete() {
	c.element.Destroy()
}

func (c *ElementComponent) Render() Instance {
	return Instance{
		element:  c.element,
		children: c.Properties.Children(),
	}
}
