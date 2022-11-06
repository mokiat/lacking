package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

var (
	TabbarHeight      = 40
	TabbarItemSpacing = 2
	TabbarSidePadding = 5
)

// Tabbar is a container intended to hold Tab components.
var Tabbar = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		layoutData = co.GetOptionalLayoutData(props, LayoutData{})
	)

	// force specific height
	layoutData.Height = opt.V(TabbarHeight)

	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: opt.V(PrimaryLightColor),
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
