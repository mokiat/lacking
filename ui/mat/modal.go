package mat

import (
	"github.com/mokiat/gog/opt"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

// Modal is a helper container that can be used as an overlay. It creates
// a dimmed background and central container.
var Modal = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: opt.V(ModalOverlayColor),
			Layout:          layout.Anchor(),
		})

		co.WithChild("content", co.New(Paper, func() {
			co.WithLayoutData(props.LayoutData())
			co.WithData(PaperData{
				Layout: layout.Frame(),
			})
			co.WithChildren(props.Children())
		}))
	})
})
