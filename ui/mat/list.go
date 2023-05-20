package mat

import (
	"github.com/mokiat/gog/opt"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	ListItemSpacing = 2
)

// List represents a container that holds a sequence of ListItem
// components in a vertical orientation.
var List = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: opt.V(SurfaceColor),
			Layout: layout.Vertical(layout.VerticalSettings{
				ContentAlignment: layout.HorizontalAlignmentLeft,
				ContentSpacing:   ListItemSpacing,
			}),
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})
