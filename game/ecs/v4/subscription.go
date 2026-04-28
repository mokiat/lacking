package ecs

import "github.com/mokiat/lacking/util/observer"

// EntityCallback represents a callback function related to an entity event.
type EntityCallback func(EntityID)

// EntitySubscription represents a notification subscription for entity events.
type EntitySubscription = observer.Subscription[EntityCallback]
