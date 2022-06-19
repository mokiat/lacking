package mvc

import (
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/filter"
)

// UseBinding is a render hook that binds the current Component to the specified
// Observable. Whenever a Change is reported by the Observable and is accepted
// by all of the specified ChangeFilters, the Component is invalidated and
// eventually re-rendered.
func UseBinding(obs Observable, fltrs ...ChangeFilter) {
	lifecycle := co.UseLifecycle(func(handle co.LifecycleHandle) *bindingLifecycle {
		return &bindingLifecycle{
			obs:    obs,
			fltr:   filter.All(fltrs...),
			handle: handle,
		}
	})
	lifecycle.ChangeObservable(obs)
}

type bindingLifecycle struct {
	co.BaseLifecycle
	obs          Observable
	fltr         ChangeFilter
	handle       co.LifecycleHandle
	subscription Subscription
}

func (l *bindingLifecycle) OnCreate(props co.Properties, scope co.Scope) {
	l.subscribe()
}

func (l *bindingLifecycle) OnDestroy(scope co.Scope) {
	l.unsubscribe()
}

func (l *bindingLifecycle) ChangeObservable(obs Observable) {
	if obs != l.obs {
		l.unsubscribe()
		l.obs = obs
		l.subscribe()
	}
}

func (l *bindingLifecycle) subscribe() {
	if l.obs != nil {
		l.subscription = l.obs.Subscribe(func(change Change) {
			l.handle.NotifyChanged()
		}, l.fltr)
	}
}

func (l *bindingLifecycle) unsubscribe() {
	if l.subscription != nil {
		l.subscription.Delete()
		l.subscription = nil
	}
}
