package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	TabbarHeight      = 40
	TabbarItemSpacing = 2
	TabbarSidePadding = 5
)

var Tabbar = co.Define(&tabbarComponent{})

type tabbarComponent struct {
	co.BaseComponent
}

func (c *tabbarComponent) Render() co.Instance {
	// force specific height
	layoutData := co.GetOptionalLayoutData(c.Properties(), layout.Data{})
	layoutData.Height = opt.V(TabbarHeight)

	return co.New(Container, func() {
		co.WithLayoutData(layoutData)
		co.WithData(ContainerData{
			BackgroundColor: opt.V(PrimaryLightColor),
			Padding: ui.Spacing{
				Left:  TabbarSidePadding,
				Right: TabbarSidePadding,
			},
			Layout: layout.Horizontal(layout.HorizontalSettings{
				ContentAlignment: layout.VerticalAlignmentCenter,
				ContentSpacing:   TabbarItemSpacing,
			}),
		})
		co.WithChildren(c.Properties().Children())
	})
}
