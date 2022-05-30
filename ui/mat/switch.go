package mat

import (
	"fmt"

	co "github.com/mokiat/lacking/ui/component"
)

type SwitchData struct {
	VisibleChildIndex int
}

var Switch = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var data SwitchData
	props.InjectOptionalData(&data, SwitchData{})

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
})
