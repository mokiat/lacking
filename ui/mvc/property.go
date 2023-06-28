package mvc

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/log"
)

var PropertyChange = NewChange("property")

func NewProperty[T any](value T) *Property[T] {
	return &Property[T]{
		value:         value,
		subscriptions: ds.NewList[*propertySubscription[T]](0),
	}
}

type Property[T any] struct {
	value         T
	subscriptions *ds.List[*propertySubscription[T]]
}

func (p *Property[T]) Get() T {
	return p.value
}

func (p *Property[T]) Set(value T) {
	p.value = value
	for _, sub := range p.subscriptions.Items() {
		sub.callback(PropertyChange)
	}
}

func (p *Property[T]) Subscribe(callback Callback) Subscription {
	log.Info("SUBSCRIBE")
	sub := &propertySubscription[T]{
		property: p,
		callback: callback,
	}
	p.subscriptions.Add(sub)
	return sub
}

type propertySubscription[T any] struct {
	property *Property[T]
	callback Callback
}

func (s *propertySubscription[T]) Delete() {
	log.Info("UNSUBSCRIBE")
	s.property.subscriptions.Remove(s)
}
