package component

import "time"

var renderCtx renderContext

type renderContext struct {
	node         *componentNode
	firstRender  bool
	lastRender   bool
	forcedRender bool
	stateIndex   int
	stateDepth   int
	properties   Properties
}

func (c renderContext) isFirstRender() bool {
	return c.firstRender
}

func (c renderContext) isLastRender() bool {
	return c.lastRender
}

// Once can be used to perform an initialization action in a
// component's render function. During subsequent renders for the
// same component instance, the specified closure function will not
// be called.
//
// You can use Once multiple times within a component's render function
// and all closure functions will be called in the respective order.
func Once(fn func()) {
	if renderCtx.isFirstRender() {
		fn()
	}
}

// Defer can be used to perform a cleanup action. The framework will issue
// one final render of a component before it gets destroyed. During that
// final render all closure functions specified via the Defer function
// will be invoked in the respective order.
//
// Similar to Once, Defer can be used multiple times within a component's
// render function.
func Defer(fn func()) {
	if renderCtx.isLastRender() {
		fn()
	}
}

// Schedule will schedule a closure function to run as soon as possible
// on the UI thread.
//
// Normally this would be used when a certain processing is being performed
// on a separate go routine and the result needs to be passed back to the
// UI thread.
//
// The framework ensures that the closure will not be called if the
// component had been destroyed in the meantime.
//
// Deprecated: This does not resolve the correct renderCtx, since it is called
// from within a go routine.
func Schedule(fn func()) {
	node := renderCtx.node
	rootUIContext.Schedule(func() {
		if node.isValid() {
			fn()
		}
	})
}

// After will schedule a closure function to be run after the specified
// amount of time. The closure is guaranteed to run on the UI thread and
// the framework ensures that the closure will not be called if the
// component had been destroyed in the meantime.
//
// Normally, you would use this function within a Once block or as a result
// of a callback.
// Not doing so would cause the closure function to be scheduled on every
// rendering of the component, since the framework is free to render a component
// at any time it deems necessary.
func After(duration time.Duration, fn func()) {
	node := renderCtx.node
	time.AfterFunc(duration, func() {
		rootUIContext.Schedule(func() {
			if node.isValid() {
				fn()
			}
		})
	})
}
