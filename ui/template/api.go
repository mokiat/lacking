package template

import (
	"fmt"
	"log"

	"github.com/mokiat/lacking/ui"
)

func New(component Component, setupFn func()) Instance {
	dslCtx = &dslContext{
		parent:          dslCtx,
		shouldReconcile: dslCtx.shouldReconcile,
	}
	defer func() {
		dslCtx = dslCtx.parent
	}()

	log.Println("NEW: ", component.componentType)
	dslCtx.instance = Instance{
		componentType: component.componentType,
		componentFunc: component.componentFunc,
	}
	setupFn()
	return dslCtx.instance
}

func WithData(data interface{}) {
	dslCtx.instance.data = data
}

func WithLayoutData(layoutData interface{}) {
	dslCtx.instance.layoutData = layoutData
}

func WithCallbackData(callbackData interface{}) {
	dslCtx.instance.callbackData = callbackData
}

func WithChild(key string, instance Instance) {
	instance.key = key
	dslCtx.instance.children = append(dslCtx.instance.children, instance)
}

func WithChildren(children []Instance) {
	dslCtx.instance.children = children
}

func Once(fn func()) {
	if renderCtx.isFirstRender() {
		fn()
	}
}

func Defer(fn func()) {
	if renderCtx.isLastRender() {
		fn()
	}
}

func UseState(fn func() interface{}) *State {
	if renderCtx.firstRender {
		renderCtx.node.states = append(renderCtx.node.states, State{
			node:  renderCtx.node,
			value: fn(),
		})
	}
	result := &renderCtx.node.states[renderCtx.stateIndex]
	renderCtx.stateIndex++
	return result
}

func OpenImage(uri string) ui.Image {
	img, err := uiCtx.OpenImage(uri)
	if err != nil {
		panic(fmt.Errorf("failed to open image %q: %w", uri, err))
	}
	return img
}
