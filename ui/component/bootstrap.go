package component

import (
	"fmt"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/layout"
)

type applicationKey struct{}

var rootUIContext *ui.Context // TODO: Remove

// Initialize wires the framework to the specified ui.Window.
// The specified Instance will be the root component used.
func Initialize(window *ui.Window, instance Instance) {
	rootUIContext = window.Context()
	rootScope := ContextScope(nil, rootUIContext)

	rootNode := createComponentNode(New(application, func() {
		WithScope(rootScope)
		WithChild("root", instance)
	}), nil)
	window.Root().AppendChild(rootNode.element)
}

var application = DefineType(&applicationComponent{})

type applicationComponent struct {
	Scope      Scope      `co:"scope"`
	Properties Properties `co:"properties"`
	Invalidate func()     `co:"invalidate"`

	childrenScope Scope
	overlays      *ds.List[*overlayHandle]
	freeOverlayID int
}

func (c *applicationComponent) OnCreate() {
	c.overlays = ds.NewList[*overlayHandle](2)
	c.childrenScope = ValueScope(c.Scope, applicationKey{}, c)
}

func (c *applicationComponent) Render() Instance {
	return New(Element, func() {
		WithData(ElementData{
			Essence: c,
			Layout:  layout.Fill(),
		})
		WithScope(c.childrenScope)
		for _, child := range c.Properties.Children() {
			WithChild(child.Key(), child)
		}
		for _, overlay := range c.overlays.Items() {
			WithChild(overlay.key, overlay.instance)
		}
	})
}

func (c *applicationComponent) OpenOverlay(instance Instance) *overlayHandle {
	c.freeOverlayID++
	result := &overlayHandle{
		app:      c,
		instance: instance,
		key:      fmt.Sprintf("overlay-%d", c.freeOverlayID),
	}
	c.overlays.Add(result)
	c.Invalidate()
	return result
}

func (c *applicationComponent) CloseOverlay(overlay *overlayHandle) {
	if !c.overlays.Remove(overlay) {
		log.Warn("[component] Overlay already closed!")
		return
	}
	c.Invalidate()
}
