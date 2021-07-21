package mat

import (
	"fmt"

	co "github.com/mokiat/lacking/ui/component"
)

type SwitchData struct {
	VisibleChildIndex int
}

var Switch = co.ShallowCached(co.Define(func(props co.Properties) co.Instance {
	var data SwitchData
	props.InjectData(&data)

	return co.New(Container, func() {
		co.WithData(ContainerData{
			Layout: NewFillLayout(),
		})
		co.WithLayoutData(props.LayoutData())
		if (0 <= data.VisibleChildIndex) && (data.VisibleChildIndex < len(props.Children())) {
			co.WithChild(
				fmt.Sprintf("visible-child-%d", data.VisibleChildIndex),
				props.Children()[data.VisibleChildIndex],
			)
		}
	})
}))
