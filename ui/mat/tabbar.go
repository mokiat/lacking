package mat

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

var (
	TabbarHeight      = 40
	TabbarItemSpacing = 2
	TabbarSidePadding = 5
)

// Tabbar is a container intended to hold Tab components.
var Tabbar = co.Define(func(props co.Properties) co.Instance {
	var (
		layoutData = co.GetOptionalLayoutData(props, LayoutData{})
	)

	// force specific height
	layoutData.Height = optional.Value(TabbarHeight)

	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: optional.Value(PrimaryLightColor),
			Padding: ui.Spacing{
				Left:  TabbarSidePadding,
				Right: TabbarSidePadding,
			},
			Layout: NewHorizontalLayout(HorizontalLayoutSettings{
				ContentAlignment: AlignmentCenter,
				ContentSpacing:   TabbarItemSpacing,
			}),
		})

		co.WithLayoutData(layoutData)
		co.WithChildren(props.Children())
	})
})
