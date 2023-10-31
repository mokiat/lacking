package game

import (
	"slices"
	"time"
)

type UpdateCallback func(elapsedTime time.Duration)

type UpdateSubscription = Subscription[UpdateCallback]

func NewSubscriptionSet[T any]() *SubscriptionSet[T] {
	return &SubscriptionSet[T]{}
}

type SubscriptionSet[T any] struct {
	subscriptions []*Subscription[T]
}

func (s *SubscriptionSet[T]) Subscribe(callback T) *Subscription[T] {
	sub := &Subscription[T]{
		set:      s,
		callback: callback,
	}
	s.subscriptions = append(s.subscriptions, sub)
	return sub
}

func (s *SubscriptionSet[T]) Unsubscribe(subscription *Subscription[T]) {
	if index := slices.Index(s.subscriptions, subscription); index >= 0 {
		s.subscriptions = slices.Delete(s.subscriptions, index, index+1)
	}
}

func (s *SubscriptionSet[T]) Each(fn func(callback T)) {
	for _, subscription := range s.subscriptions {
		fn(subscription.callback)
	}
}

func (s *SubscriptionSet[T]) Clear() {
	s.subscriptions = nil
}

type Subscription[T any] struct {
	set      *SubscriptionSet[T]
	callback T
}

func (s *Subscription[T]) Delete() {
	s.set.Unsubscribe(s)
}
