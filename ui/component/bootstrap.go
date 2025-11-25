package component

import (
	"fmt"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/ui/layout"
)

// Initialize wires the framework using the specified root scope.
// The specified instance will be the root component used.
func Initialize(scope Scope, instance Instance) {
	window := Window(scope)

	element := window.CreateElement()
	window.Root().AppendChild(element)

	createComponentNode(element, scope, New(application, func() {
		WithChild("root", instance)
	}))
}

var application = Define[*applicationComponent]()

type applicationComponent struct {
	BaseComponent

	overlays      *ds.List[*overlayHandle]
	freeOverlayID int
}

func (c *applicationComponent) OnCreate() {
	c.overlays = ds.NewList[*overlayHandle](2)
}

func (c *applicationComponent) Render() Instance {
	return New(Element, func() {
		WithData(ElementData{
			Essence: c,
			Layout:  layout.Fill(),
		})
		WithTypedScopeValue(c)
		for _, child := range c.Properties().Children() {
			WithChild(child.Key(), child)
		}
		for _, overlay := range c.overlays.Unbox() {
			WithChild(overlay.instance.key, overlay.instance)
		}
	})
}

func (c *applicationComponent) openOverlay(instance Instance) *overlayHandle {
	result := &overlayHandle{
		app: c,
	}

	c.freeOverlayID++
	instance.key = fmt.Sprintf("overlay-%d", c.freeOverlayID)
	instance.addScopeModifier(func(scope Scope) Scope {
		return TypedValueScope(scope, result)
	})
	result.instance = instance

	c.overlays.Add(result)
	c.Invalidate()
	return result
}

func (c *applicationComponent) closeOverlay(overlay *overlayHandle) {
	if !c.overlays.Remove(overlay) {
		logger.Warn("Overlay already closed")
	}
	c.Invalidate()
}
