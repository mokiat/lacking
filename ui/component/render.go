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
func After(scope Scope, duration time.Duration, fn func()) {
	time.AfterFunc(duration, func() {
		Schedule(scope, fn)
	})
}

// Every will schedule a closure function to be run every interval amount of
// time. The closure is guaranteed to run on the UI thread and
// the framework ensures that the closure will not be called if the
// component had been destroyed in the meantime. In fact, the closure will
// stop being called only once the component has been destroyed.
func Every(scope Scope, interval time.Duration, fn func()) {
	time.AfterFunc(interval, func() {
		fn()
		Every(scope, interval, fn)
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
