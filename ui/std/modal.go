package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var Modal = co.Define(&modalComponent{})

type modalComponent struct {
	co.BaseComponent
}

func (c *modalComponent) Render() co.Instance {
	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence:   c,
			Layout:    layout.Fill(),
			Focusable: opt.V(true),
			Focused:   opt.V(true),
		})

		co.WithChild("shading", co.New(Container, func() {
			co.WithData(ContainerData{
				BackgroundColor: opt.V(ModalOverlayColor),
				Layout:          layout.Anchor(),
			})

			co.WithChild("content", co.New(Paper, func() {
				co.WithLayoutData(c.Properties().LayoutData())
				co.WithData(PaperData{
					Layout: layout.Frame(),
				})
				co.WithChildren(c.Properties().Children())
			}))
		}))
	})
}

var _ ui.ElementKeyboardHandler = (*modalComponent)(nil)
var _ ui.ElementMouseHandler = (*modalComponent)(nil)

func (c *modalComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	// Prevet lower layers from accessing key events.
	return true
}

func (c *modalComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	// Prevet lower layers from accessing mouse events.
	return true
}
