package timestep

import (
	"slices"
	"time"
)

// UpdateCallback is a mechanism to receive delta time increments.
type UpdateCallback func(elapsedTime time.Duration)

// UpdateSubscription represents a notification subscription for updates.
type UpdateSubscription = Subscription[UpdateCallback]

// NewSubscriptionSet creates a new SubscriptionSet.
func NewSubscriptionSet[T any]() *SubscriptionSet[T] {
	return &SubscriptionSet[T]{}
}

// SubscriptionSet holds a list of subscribers.
type SubscriptionSet[T any] struct {
	subscriptions []*Subscription[T]
}

// Subscribe adds a subscription to this set.
func (s *SubscriptionSet[T]) Subscribe(callback T) *Subscription[T] {
	sub := &Subscription[T]{
		set:      s,
		callback: callback,
	}
	s.subscriptions = append(s.subscriptions, sub)
	return sub
}

// Unsubscribe removes the subscription from this set.
func (s *SubscriptionSet[T]) Unsubscribe(subscription *Subscription[T]) {
	if index := slices.Index(s.subscriptions, subscription); index >= 0 {
		s.subscriptions = slices.Delete(s.subscriptions, index, index+1)
	}
}

// Each iterates over all subscriptions.
func (s *SubscriptionSet[T]) Each(fn func(callback T)) {
	for _, subscription := range s.subscriptions {
		fn(subscription.callback)
	}
}

// Clear removes all subscriptions.
func (s *SubscriptionSet[T]) Clear() {
	s.subscriptions = nil
}

// Subscription represents a registration for notifications.
type Subscription[T any] struct {
	set      *SubscriptionSet[T]
	callback T
}

// Delete removes the subscription from what it is watching.
func (s *Subscription[T]) Delete() {
	s.set.Unsubscribe(s)
}
