package mat

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

var defaultSwitchData = SwitchData{}

// Switch is a container that can switch between different views depending
// on the specified SwitchData.
var Switch = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data = co.GetOptionalData(props, defaultSwitchData)
	)

	return co.New(Element, func() {
		co.WithData(ElementData{
			Layout: layout.Fill(),
		})
		co.WithLayoutData(props.LayoutData())

		for _, child := range props.Children() {
			if child.Key() == data.ChildKey {
				co.WithChild(child.Key(), child)
			}
		}
	})
})
