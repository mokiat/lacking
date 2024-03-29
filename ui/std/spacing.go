package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

// SpacingData holds the configuration data for a Spacing component.
type SpacingData struct {
	Size ui.Size
}

var spacingDefaultData = SpacingData{
	Size: ui.NewSize(5, 5),
}

// Spacing represents an invisible component that requests that a specific
// amount of visual space be reserved for it.
var Spacing = co.Define(&spacingComponent{})

type spacingComponent struct {
	co.BaseComponent

	spacing ui.Size
}

func (c *spacingComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), spacingDefaultData)
	c.spacing = data.Size
}

func (c *spacingComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			IdealSize: opt.V(c.spacing),
		})
	})
}
