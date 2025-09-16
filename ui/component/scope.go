package component

import (
	"time"

	"github.com/mokiat/lacking/ui"
)

// Scope represents a component sub-hierarchy region.
type Scope interface {

	// Context returns the ui.Context that is applicable for this component scope.
	Context() *ui.Context

	// Value returns the stored arbitrary value for the specified arbitrary key.
	// This is a mechanism through which external frameworks can attach metadata
	// to scopes.
	Value(key any) any
}

// RootScope initializes a new scope associated with the specified window.
//
// One would usually use this method to acquire a root scope to be later
// used in Initialize to bootstrap the framework.
func RootScope(window *ui.Window) Scope {
	return ContextScope(nil, window.Context())
}

// Window uses the specified scope to retrieve the Window that owns that
// particular scope.
func Window(scope Scope) *ui.Window {
	return scope.Context().Window()
}

// Invalidate causes the component that owns the specified scope to be
// reconciled.
//
// This function should only be called from the UI thread.
func Invalidate(scope Scope) {
	node := componentNodeFromScope(scope)
	if node.isValid() {
		node.invalidate()
	}
}

// Schedule will schedule a closure function to run as soon as possible
// on the UI thread. It is safe to call this function from any thread.
//
// Normally this would be used when a certain processing is being performed
// on a separate go routine and the result needs to be passed back to the
// UI thread.
//
// The framework ensures that the closure will not be called if the
// component had been destroyed in the meantime.
func Schedule(scope Scope, fn func()) {
	scope.Context().Schedule(func() {
		node := componentNodeFromScope(scope)
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
// It is safe to call this function from any thread.
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
//
// It is safe to call this function from any thread.
func Every(scope Scope, interval time.Duration, fn func()) {
	After(scope, interval, func() {
		fn()
		Every(scope, interval, fn)
	})
}
