package ecs

import "github.com/mokiat/lacking/util/observer"

// EntityCallback is a function invoked when an entity event fires.
// The id parameter identifies the entity that triggered the event.
type EntityCallback func(ID)

type ConditionalCallback struct {
	condition Condition
	callback  EntityCallback
}

// EntitySubscription is a handle to a registered enter or exit
// callback. Call its Delete method to cancel the subscription.
type EntitySubscription = observer.Subscription[ConditionalCallback]
