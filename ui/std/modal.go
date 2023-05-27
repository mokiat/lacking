package std

import (
	"github.com/mokiat/gog/opt"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var Modal = co.Define(&modalComponent{})

type modalComponent struct {
	Properties co.Properties `co:"properties"`
}

func (c *modalComponent) Render() co.Instance {
	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: opt.V(ModalOverlayColor),
			Layout:          layout.Anchor(),
		})

		co.WithChild("content", co.New(Paper, func() {
			co.WithLayoutData(c.Properties.LayoutData())
			co.WithData(PaperData{
				Layout: layout.Frame(),
			})
			co.WithChildren(c.Properties.Children())
		}))
	})
}
