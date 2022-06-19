package mat

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

// SpacingData holds the configuration data for a Spacing component.
type SpacingData struct {
	Width  int
	Height int
}

var defaultSpacingData = SpacingData{}

// Spacing represents a non-visual component that just takes up
// space and is intended to be used as a separator in layouts.
var Spacing = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	data := co.GetOptionalData(props, defaultSpacingData)
	return co.New(Element, func() {
		co.WithLayoutData(props.LayoutData())
		co.WithData(ElementData{
			IdealSize: optional.Value(ui.Size{
				Width:  data.Width,
				Height: data.Height,
			}),
		})
	})
})
