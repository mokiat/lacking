package component

// UseLifecycle attaches a Lifecycle to the given Component instance in which
// this hook is used.
func UseLifecycle[T Lifecycle](constructor func(handle LifecycleHandle) T) T {
	changeState := UseState(func() interface{} {
		return int(0)
	})

	var instance T
	UseState(func() interface{} {
		return constructor(LifecycleHandle{
			changeState: changeState,
		})
	}).Inject(&instance)

	switch {
	case renderCtx.firstRender:
		instance.OnCreate(renderCtx.properties)
	case renderCtx.lastRender:
		instance.OnDestroy()
	default:
		instance.OnUpdate(renderCtx.properties)
	}
	return instance
}

// LifecycleHandle is a hook through which a Lifecycle can invalidate a
// Component, as though the Component's data were changed.
type LifecycleHandle struct {
	changeState *State
}

// NotifyChanged invalidates the Component instance.
func (h LifecycleHandle) NotifyChanged() {
	h.changeState.Set(h.changeState.Get().(int) + 1)
}

// Lifecycle is a mechanism through which a Component may get more detailed
// lifecycle events.
type Lifecycle interface {

	// OnCreate is called when the Component instance is first created.
	OnCreate(props Properties)

	// OnUpdate is called when an existing Component instance has its properties
	// changed.
	OnUpdate(props Properties)

	// OnDestroy is called when the Component is about to be detached and removed.
	OnDestroy()
}

// NewBaseLifecycle returns a new BaseLifecycle.
func NewBaseLifecycle() *BaseLifecycle {
	return &BaseLifecycle{}
}

var _ Lifecycle = (*BaseLifecycle)(nil)

// BaseLifecycle is an implementation of Lifecycle that does nothing.
type BaseLifecycle struct{}

func (l *BaseLifecycle) OnCreate(props Properties) {}

func (l *BaseLifecycle) OnUpdate(props Properties) {}

func (l *BaseLifecycle) OnDestroy() {}
