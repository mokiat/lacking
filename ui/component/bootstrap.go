package component

import (
	"fmt"

	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/ui"
	"golang.org/x/exp/slices"
)

var rootLifecycle *windowLifecycle

// Initialize wires the framework to the specified ui Window.
// The specified instance will be the root component used.
func Initialize(window *ui.Window, instance Instance) {
	// TODO: Destroy uiCtx at the end.
	uiCtx = window.Context()
	rootScope = ContextScope(nil, uiCtx)

	rootNode := createComponentNode(New(application, func() {
		WithScope(rootScope)
		WithChild("root", instance)
	}), nil)
	window.Root().AppendChild(rootNode.element)
}

var application = Define(func(props Properties, scope Scope) Instance {
	lifecycle := UseLifecycle(func(handle LifecycleHandle) *windowLifecycle {
		return &windowLifecycle{
			BaseLifecycle: NewBaseLifecycle(),
			handle:        handle,
		}
	})
	return New(Element, func() {
		WithData(ElementData{
			Essence: lifecycle,
			Layout:  ui.NewFillLayout(),
		})
		for _, child := range props.Children() {
			WithChild(child.Key(), child)
		}
		for _, overlay := range lifecycle.overlays {
			WithChild(overlay.key, overlay.instance)
		}
	})
})

type windowLifecycle struct {
	*BaseLifecycle
	handle        LifecycleHandle
	overlays      []*overlayHandle
	freeOverlayID int
}

func (l *windowLifecycle) OnCreate(props Properties, scope Scope) {
	rootLifecycle = l
}

func (l *windowLifecycle) OnDestroy(scope Scope) {
	rootLifecycle = nil
}

func (l *windowLifecycle) OpenOverlay(instance Instance) *overlayHandle {
	l.freeOverlayID++
	result := &overlayHandle{
		lifecycle: l,
		instance:  instance,
		key:       fmt.Sprintf("overlay-%d", l.freeOverlayID),
	}
	l.overlays = append(l.overlays, result)
	l.handle.NotifyChanged()
	return result
}

func (l *windowLifecycle) CloseOverlay(overlay *overlayHandle) {
	index := slices.Index(l.overlays, overlay)
	if index < 0 {
		log.Warn("[component] Overlay already closed")
		return
	}
	l.overlays = slices.Delete(l.overlays, index, index+1)
	l.handle.NotifyChanged()
}
