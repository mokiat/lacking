package component

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

var bootstrapCtrl *bootstrapController

// Initialize wires the framework to the specified ui Window.
// The specified instance will be the root component used.
func Initialize(window *ui.Window, instance Instance) {
	uiCtx = window.Context()

	bootstrapCtrl = &bootstrapController{
		Controller: NewBaseController(),
		root:       instance,
	}
	rootNode := createComponentNode(New(application, func() {
		WithData(bootstrapCtrl)
	}))
	window.Root().AppendChild(rootNode.element)
}

var application = Controlled(Define(func(props Properties) Instance {
	controller := props.Data().(*bootstrapController)
	return New(Element, func() {
		WithData(ElementData{
			Layout: ui.NewFillLayout(),
		})
		WithChild("root", controller.root)
		for _, overlay := range controller.overlays {
			WithChild(fmt.Sprintf("overlay-%d", overlay.id), overlay.instance)
		}
	})
}))

func OpenOverlay(instance Instance) Overlay {
	return Overlay{
		id: bootstrapCtrl.AddOverlay(instance),
	}
}

type Overlay struct {
	id int
}

func (o *Overlay) Close() {
	bootstrapCtrl.RemoveOverlay(o.id)
	o.id = -1
}

type bootstrapController struct {
	Controller
	root          Instance
	overlays      []appOverlay
	overlayFreeID int
}

func (c *bootstrapController) AddOverlay(instance Instance) int {
	c.overlayFreeID++
	c.overlays = append(c.overlays, appOverlay{
		id:       c.overlayFreeID,
		instance: instance,
	})
	c.NotifyChanged()
	return c.overlayFreeID
}

func (c *bootstrapController) RemoveOverlay(id int) {
	index := c.findOverlay(id)
	if index < 0 {
		panic("overlay already removed!")
	}
	c.overlays = append(c.overlays[:index], c.overlays[index+1:]...)
	c.NotifyChanged()
}

func (c *bootstrapController) findOverlay(id int) int {
	for i, overlay := range c.overlays {
		if overlay.id == id {
			return i
		}
	}
	return -1
}

type appOverlay struct {
	id       int
	instance Instance
}
