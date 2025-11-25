package ecs

import "github.com/mokiat/lacking/util/observer"

// DeleteCallback is a mechanism to receive deletion notifications.
type DeleteCallback func(EntityID)

// DeleteSubscription represents a notification subscription for deletions.
type DeleteSubscription = observer.Subscription[DeleteCallback]
