package ecs

import "github.com/mokiat/lacking/util/observer"

type EntityCallback func(ID)

type ConditionalCallback struct {
	condition Condition
	callback  EntityCallback
}

type EntitySubscription = observer.Subscription[ConditionalCallback]
