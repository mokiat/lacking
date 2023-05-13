package component

import (
	"fmt"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/ui"
)

type applicationKey struct{}

var (
	rootUIContext *ui.Context
	rootScope     Scope
)

// Initialize wires the framework to the specified ui.Window.
// The specified Instance will be the root component used.
func Initialize(window *ui.Window, instance Instance) {
	rootUIContext = window.Context()
	rootScope = ContextScope(nil, rootUIContext)

	rootNode := createComponentNode(New(application, func() {
		WithScope(rootScope)
		WithChild("root", instance)
	}), nil)
	window.Root().AppendChild(rootNode.element)
}

// TODO: Use TypeDefine instead
var application = Define(func(props Properties, scope Scope) Instance {
	lifecycle := UseLifecycle(func(handle LifecycleHandle) *windowLifecycle {
		return &windowLifecycle{
			BaseLifecycle: NewBaseLifecycle(),
			overlays:      ds.NewList[*overlayHandle](0),
			handle:        handle,
		}
	})
	return New(Element, func() {
		WithData(ElementData{
			Essence: lifecycle,
			Layout:  ui.NewFillLayout(),
		})
		WithScope(ValueScope(scope, applicationKey{}, lifecycle)) // TODO: cache
		for _, child := range props.Children() {
			WithChild(child.Key(), child)
		}
		for _, overlay := range lifecycle.overlays.Items() {
			WithChild(overlay.key, overlay.instance)
		}
	})
})

type windowLifecycle struct {
	*BaseLifecycle
	handle        LifecycleHandle
	overlays      *ds.List[*overlayHandle]
	freeOverlayID int
}

func (l *windowLifecycle) OpenOverlay(instance Instance) *overlayHandle {
	l.freeOverlayID++
	result := &overlayHandle{
		lifecycle: l,
		instance:  instance,
		key:       fmt.Sprintf("overlay-%d", l.freeOverlayID),
	}
	l.overlays.Add(result)
	l.handle.NotifyChanged()
	return result
}

func (l *windowLifecycle) CloseOverlay(overlay *overlayHandle) {
	if !l.overlays.Remove(overlay) {
		log.Warn("[component] Overlay already closed!")
		return
	}
	l.handle.NotifyChanged()
}
