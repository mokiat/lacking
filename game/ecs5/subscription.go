package ecs5

import "github.com/mokiat/lacking/util/observer"

// DeleteCallback is a mechanism to receive deletion notifications.
type DeleteCallback func(entity Entity)

// DeleteSubscription represents a notification subscription for deletions.
type DeleteSubscription = observer.Subscription[DeleteCallback]
