package component

// UseLifecycle attaches a Lifecycle to the given Component instance in which
// this hook is used.
//
// Deprecated: Use new component types
func UseLifecycle[T Lifecycle](constructor func(handle LifecycleHandle) T) T {
	changeState := UseState(func() int {
		return int(0)
	})

	instance := UseState(func() T {
		return constructor(LifecycleHandle{
			changeState: changeState,
		})
	}).Get()

	switch {
	case renderCtx.firstRender:
		instance.OnCreate(renderCtx.properties, renderCtx.node.scope)
	case renderCtx.lastRender:
		instance.OnDestroy(renderCtx.node.scope)
	default:
		instance.OnUpdate(renderCtx.properties, renderCtx.node.scope)
	}
	return instance
}

// LifecycleHandle is a hook through which a Lifecycle can invalidate a
// Component, as though the Component's data were changed.
type LifecycleHandle struct {
	changeState *State[int]
}

// NotifyChanged invalidates the Component instance.
func (h LifecycleHandle) NotifyChanged() {
	h.changeState.Set(h.changeState.Get() + 1)
}

// Lifecycle is a mechanism through which a Component may get more detailed
// lifecycle events.
type Lifecycle interface {

	// OnCreate is called when the Component instance is first created.
	OnCreate(props Properties, scope Scope)

	// OnUpdate is called when an existing Component instance has its properties
	// changed.
	OnUpdate(props Properties, scope Scope)

	// OnDestroy is called when the Component is about to be detached and removed.
	OnDestroy(scope Scope)
}

// NewBaseLifecycle returns a new BaseLifecycle.
//
// Deprecated: Use new component types
func NewBaseLifecycle() *BaseLifecycle {
	return &BaseLifecycle{}
}

var _ Lifecycle = (*BaseLifecycle)(nil)

// BaseLifecycle is an implementation of Lifecycle that does nothing.
type BaseLifecycle struct{}

func (l *BaseLifecycle) OnCreate(props Properties, scope Scope) {}

func (l *BaseLifecycle) OnUpdate(props Properties, scope Scope) {}

func (l *BaseLifecycle) OnDestroy(scope Scope) {}
