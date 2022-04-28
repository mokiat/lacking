package component

type LifecycleHandle struct {
	changeState *State
}

func (h LifecycleHandle) NotifyChanged() {
	h.changeState.Set(h.changeState.Get().(int) + 1)
}

func UseLifecycle(constructor func(handle LifecycleHandle) Lifecycle, target ...interface{}) {
	changeState := UseState(func() interface{} {
		return int(0)
	})

	var instance Lifecycle
	UseState(func() interface{} {
		return constructor(LifecycleHandle{
			changeState: changeState,
		})
	}).Inject(&instance)

	if len(target) > 0 {
		inject(target[0], instance)
	}

	switch {
	case renderCtx.firstRender:
		instance.OnCreate(renderCtx.properties)
	case renderCtx.lastRender:
		instance.OnDestroy()
	default:
		instance.OnUpdate(renderCtx.properties)
	}
}

type Lifecycle interface {
	OnCreate(props Properties)
	OnUpdate(props Properties)
	OnDestroy()
}

func NewBaseLifecycle() Lifecycle {
	return &BaseLifecycle{}
}

var _ Lifecycle = (*BaseLifecycle)(nil)

type BaseLifecycle struct {
}

func (l *BaseLifecycle) OnCreate(props Properties) {}

func (l *BaseLifecycle) OnUpdate(props Properties) {}

func (l *BaseLifecycle) OnDestroy() {}
