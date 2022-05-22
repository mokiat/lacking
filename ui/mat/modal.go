package mat

import (
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

// Modal is a helper container that can be used as an overlay. It creates
// a dimmed background and central container.
var Modal = co.Define(func(props co.Properties) co.Instance {
	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: optional.Value(ModalOverlayColor),
			Layout:          NewAnchorLayout(AnchorLayoutSettings{}),
		})

		co.WithChild("content", co.New(Paper, func() {
			co.WithData(PaperData{
				Layout: NewFrameLayout(),
			})
			co.WithLayoutData(props.LayoutData())
			co.WithChildren(props.Children())
		}))
	})
})
