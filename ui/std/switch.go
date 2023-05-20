package std

import (
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

// SwitchData holds the data for a Switch component.
type SwitchData struct {

	// ChildKey holds the key of the child component to be rendered.
	// If this does not match any child then nothing is rendered.
	ChildKey string
}

var switchDefaultData = SwitchData{}

// Switch is a container that can switch between different views depending
// on the specified SwitchData.
var Switch = co.DefineType(&SwitchComponent{})

type SwitchComponent struct {
	Properties co.Properties `co:"properties"`

	activeChildKey string
}

func (c *SwitchComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties, switchDefaultData)
	c.activeChildKey = data.ChildKey
}

func (c *SwitchComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithData(co.ElementData{
			Layout: layout.Fill(),
		})
		co.WithLayoutData(c.Properties.LayoutData())

		for _, child := range c.Properties.Children() {
			if child.Key() == c.activeChildKey {
				co.WithChild(child.Key(), child)
			}
		}
	})
}
