package component

import (
	"fmt"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	coLogger = log.Path("/ui/component")
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
		for _, overlay := range c.overlays.Items() {
			WithChild(overlay.instance.key, overlay.instance)
		}
	})
}

func (c *applicationComponent) OpenOverlay(instance Instance) *overlayHandle {
	c.freeOverlayID++
	instance.key = fmt.Sprintf("overlay-%d", c.freeOverlayID)
	result := &overlayHandle{
		app:      c,
		instance: instance,
	}
	c.overlays.Add(result)
	c.Invalidate()
	return result
}

func (c *applicationComponent) CloseOverlay(overlay *overlayHandle) {
	if !c.overlays.Remove(overlay) {
		panic("closing a closed overlay")
	}
	c.Invalidate()
}
