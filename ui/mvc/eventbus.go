package mvc

import "golang.org/x/exp/slices"

// Event represents an arbitrary notification event.
type Event any

// CallbackFunc is a mechanism to be notified of an Event.
type CallbackFunc func(Event)

// NewEventBus creates a new EventBus instance.
func NewEventBus() *EventBus {
	return &EventBus{}
}

// EventBus is a mechanism to listen for global events from within your
// components.
//
// The general pattern is to have on such EventBus and inject it in the
// root Scope to be accessible by all components.
//
// Then components that need to have special invalidation logic would subscribe
// and depending on the event would call Invalidate on themselves. If the
// event is too generic as a type, then its fields need to narrow down the
// receiver as much as possible, otherwise there is a risk that too many
// components would be invalidated without need.
type EventBus struct {
	subscriptions []*eventBusSubscription
}

// Notify sends the specified event to all subscribed listeners.
func (b *EventBus) Notify(event Event) {
	for _, sub := range b.subscriptions {
		sub.callback(event)
	}
}

// Subscribe adds the specified callback to be invoked whenever an event
// occurs.
//
// The returned Subscription can be used to unregister the callback.
func (b *EventBus) Subscribe(callback CallbackFunc) Subscription {
	sub := &eventBusSubscription{
		eventBus: b,
		callback: callback,
	}
	b.subscriptions = append(b.subscriptions, sub)
	return sub
}

// Unsubscribe disables the specified subscription and future events would
// not be sent to it.
func (b *EventBus) Unsubscribe(subscription Subscription) {
	if ebSub, ok := subscription.(*eventBusSubscription); ok {
		b.subscriptions = slices.DeleteFunc(b.subscriptions, func(candidate *eventBusSubscription) bool {
			return candidate == ebSub
		})
		b.subscriptions = slices.Clip(b.subscriptions)
	}
}

type eventBusSubscription struct {
	eventBus *EventBus
	callback CallbackFunc
}

func (s *eventBusSubscription) Delete() {
	s.eventBus.Unsubscribe(s)
}
