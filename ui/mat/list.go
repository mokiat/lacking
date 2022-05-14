package mat

import (
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

var (
	ListItemSpacing = 2
)

// List represents a container that holds a sequence of ListItem
// components in a vertical orientation.
var List = co.Define(func(props co.Properties) co.Instance {
	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: optional.Value(SurfaceColor),
			Layout: NewVerticalLayout(VerticalLayoutSettings{
				ContentAlignment: AlignmentLeft,
				ContentSpacing:   ListItemSpacing,
			}),
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})
