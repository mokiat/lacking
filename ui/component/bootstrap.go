package component

import (
	"fmt"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/layout"
)

var rootUIContext *ui.Context // TODO: Remove

// Initialize wires the framework using the specified root scope.
// The specified instance will be the root component used.
func Initialize(scope Scope, instance Instance) {
	window := Window(scope)
	if instance.scope != nil {
		panic(fmt.Errorf("root instances should not have scope assigned"))
	}
	rootNode := createComponentNode(New(application, func() {
		WithScope(scope)
		WithChild("root", instance)
	}), nil)
	window.Root().AppendChild(rootNode.element)
}

var application = Define(&applicationComponent{})

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
	c.childrenScope = TypedValueScope(c.Scope, c)
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
