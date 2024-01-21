package physics

import (
	"time"

	"github.com/mokiat/lacking/util/observer"
)

// UpdateCallback is a mechanism to receive update notifications.
type UpdateCallback func(elapsedTime time.Duration)

// UpdateSubscription represents a notification subscription for updates.
type UpdateSubscription = observer.Subscription[UpdateCallback]

// UpdateSubscriptionSet represents a set of update subscriptions.
type UpdateSubscriptionSet = observer.SubscriptionSet[UpdateCallback]

// NewUpdateSubscriptionSet creates a new UpdateSubscriptionSet.
func NewUpdateSubscriptionSet() *UpdateSubscriptionSet {
	return observer.NewSubscriptionSet[UpdateCallback]()
}

type DynamicCollisionCallback func(first, second *Body, colliding bool)

type DynamicCollisionSubscription = observer.Subscription[DynamicCollisionCallback]

type DynamicCollisionSubscriptionSet = observer.SubscriptionSet[DynamicCollisionCallback]

func NewDynamicCollisionSubscriptionSet() *DynamicCollisionSubscriptionSet {
	return observer.NewSubscriptionSet[DynamicCollisionCallback]()
}

type StaticCollisionCallback func(body *Body)

type StaticCollisionSubscription = observer.Subscription[StaticCollisionCallback]

type StaticCollisionSubscriptionSet = observer.SubscriptionSet[StaticCollisionCallback]

func NewStaticCollisionSubscriptionSet() *StaticCollisionSubscriptionSet {
	return observer.NewSubscriptionSet[StaticCollisionCallback]()
}
