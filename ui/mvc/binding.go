package mvc

import (
	"fmt"

	co "github.com/mokiat/lacking/ui/component"
)

type mvcScopeKey struct{}

func Wrap(delegate co.Component) co.Component {
	return &mvcComponent{
		Component:            delegate,
		pendingSubscriptions: make(map[co.Renderable][]Subscription),
		activeSubscriptions:  make(map[co.Renderable][]Subscription),
	}
}

type mvcComponent struct {
	co.Component

	activeRef            co.Renderable
	pendingSubscriptions map[co.Renderable][]Subscription
	activeSubscriptions  map[co.Renderable][]Subscription
}

func (c *mvcComponent) TypeName() string {
	return fmt.Sprintf("mvc(%s)", c.Component.TypeName())
}

func (c *mvcComponent) Allocate(scope co.Scope, invalidate co.InvalidateFunc) co.Renderable {
	scope = co.ValueScope(scope, mvcScopeKey{}, c)
	return c.Component.Allocate(scope, invalidate)
}

func (c *mvcComponent) NotifyCreate(ref co.Renderable, properties co.Properties) {
	c.activeRef = ref
	c.Component.NotifyCreate(ref, properties)
	c.replaceSubscriptions(ref)
	c.activeRef = nil
}

func (c *mvcComponent) NotifyUpdate(ref co.Renderable, properties co.Properties) {
	c.activeRef = ref
	c.Component.NotifyUpdate(ref, properties)
	c.replaceSubscriptions(ref)
	c.activeRef = nil
}

func (c *mvcComponent) NotifyDelete(ref co.Renderable) {
	c.deleteSubscriptions(ref)
	c.Component.NotifyDelete(ref)
}

func (c *mvcComponent) addPendingSubscription(subscription Subscription) {
	ref := c.activeRef
	c.pendingSubscriptions[ref] = append(c.pendingSubscriptions[ref], subscription)
}

func (c *mvcComponent) replaceSubscriptions(ref co.Renderable) {
	if pendingSubcriptions := c.pendingSubscriptions[ref]; len(pendingSubcriptions) > 0 {
		c.deleteSubscriptions(ref)
		c.activeSubscriptions[ref] = pendingSubcriptions
		c.pendingSubscriptions[ref] = nil
	}
}

func (c *mvcComponent) deleteSubscriptions(ref co.Renderable) {
	for _, subscription := range c.activeSubscriptions[ref] {
		subscription.Delete()
	}
	c.activeSubscriptions[ref] = nil
}

// UseBinding is a hook that binds the current Component to the specified
// Observable. Whenever a Change is reported by the Observable and is accepted
// by all of the specified ChangeFilters, the Component is invalidated and
// eventually re-rendered.
func UseBinding(scope co.Scope, obs Observable, fltrs ...ChangeFilter) {
	scopeValue := scope.Value(mvcScopeKey{})
	mvcComp, ok := scopeValue.(*mvcComponent)
	if !ok {
		panic("component not wrapped with mvc")
	}
	if mvcComp.activeRef == nil {
		panic("binding not allowed here")
	}
	mvcComp.addPendingSubscription(obs.Subscribe(func(change Change) {
		co.Invalidate(scope)
	}, fltrs...))
}
