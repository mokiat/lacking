package component

import "time"

// Schedule will schedule a closure function to run as soon as possible
// on the UI thread.
//
// Normally this would be used when a certain processing is being performed
// on a separate go routine and the result needs to be passed back to the
// UI thread.
//
// The framework ensures that the closure will not be called if the
// component had been destroyed in the meantime.
func Schedule(scope Scope, fn func()) {
	node := scope.Value(componentNodeKey{}).(*componentNode)
	scope.Context().Schedule(func() {
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
func After(scope Scope, duration time.Duration, fn func()) {
	time.AfterFunc(duration, func() {
		Schedule(scope, fn)
	})
}

// Invalidate causes the component that owns the specified scope to be
// recalculated.
func Invalidate(scope Scope) {
	node := scope.Value(componentNodeKey{}).(*componentNode)
	if node.isValid() {
		node.invalidate()
	}
}
