package std

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
var List = co.DefineType(&ListComponent{})

type ListComponent struct {
	Properties co.Properties `co:"properties"`
}

func (c *ListComponent) Render() co.Instance {
	return co.New(Container, func() {
		co.WithLayoutData(c.Properties.LayoutData())
		co.WithData(ContainerData{
			BackgroundColor: opt.V(SurfaceColor),
			Layout: layout.Vertical(layout.VerticalSettings{
				ContentAlignment: layout.HorizontalAlignmentLeft,
				ContentSpacing:   ListItemSpacing,
			}),
		})
		co.WithChildren(c.Properties.Children())
	})
}
