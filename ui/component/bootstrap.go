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
	if instance.scope != nil {
		panic(fmt.Errorf("root instances should not have scope assigned"))
	}
	rootNode := createComponentNode(scope, New(application, func() {
		WithChild("root", instance)
	}))
	window.Root().AppendChild(rootNode.leafElement())
}

var application = Define(&applicationComponent{})

type applicationComponent struct {
	BaseComponent

	childrenScope Scope
	overlays      *ds.List[*overlayHandle]
	freeOverlayID int
}

func (c *applicationComponent) OnCreate() {
	c.overlays = ds.NewList[*overlayHandle](2)
	c.childrenScope = TypedValueScope(c.Scope(), c)
}

func (c *applicationComponent) Render() Instance {
	return New(Element, func() {
		WithData(ElementData{
			Essence: c,
			Layout:  layout.Fill(),
		})
		WithScope(c.childrenScope)
		for _, child := range c.Properties().Children() {
			WithChild(child.Key(), child)
		}
		for _, overlay := range c.overlays.Unbox() {
			WithChild(overlay.instance.key, overlay.instance)
		}
	})
}

func (c *applicationComponent) OpenOverlay(scope Scope, instance Instance) *overlayHandle {
	c.freeOverlayID++

	result := &overlayHandle{
		app: c,
	}

	instance.key = fmt.Sprintf("overlay-%d", c.freeOverlayID)
	instance.setScope(TypedValueScope(scope, result))
	result.instance = instance

	c.overlays.Add(result)
	c.Invalidate()
	return result
}

func (c *applicationComponent) CloseOverlay(overlay *overlayHandle) {
	if !c.overlays.Remove(overlay) {
		logger.Warn("Overlay already closed!")
	}
	c.Invalidate()
}
