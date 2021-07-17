package mat

import (
	"fmt"

	t "github.com/mokiat/lacking/ui/template"
)

type SwitchData struct {
	VisibleChildIndex int
}

var Switch = t.ShallowCached(t.Plain(func(props t.Properties) t.Instance {
	var data SwitchData
	props.InjectData(&data)

	return t.New(Container, func() {
		t.WithData(ContainerData{
			Layout: NewFillLayout(),
		})
		t.WithLayoutData(props.LayoutData())
		if (0 <= data.VisibleChildIndex) && (data.VisibleChildIndex < len(props.Children())) {
			t.WithChild(
				fmt.Sprintf("visible-child-%d", data.VisibleChildIndex),
				props.Children()[data.VisibleChildIndex],
			)
		}
	})
}))
