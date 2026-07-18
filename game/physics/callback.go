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

// DoubleBodyCollisionCallback is a mechanism to receive notifications
// about collisions between two bodies.
type DoubleBodyCollisionCallback func(first, second Body, active bool)

// DoubleBodyCollisionSubscription represents a notification subscription
// for double body collisions.
type DoubleBodyCollisionSubscription = observer.Subscription[DoubleBodyCollisionCallback]

// DoubleBodyCollisionSubscriptionSet represents a set of double body
// collision subscriptions.
type DoubleBodyCollisionSubscriptionSet = observer.SubscriptionSet[DoubleBodyCollisionCallback]

// NewDoubleBodyCollisionSubscriptionSet creates a new
// DoubleBodyCollisionSubscriptionSet.
func NewDoubleBodyCollisionSubscriptionSet() *DoubleBodyCollisionSubscriptionSet {
	return observer.NewSubscriptionSet[DoubleBodyCollisionCallback]()
}

// SingleBodyCollisionCallback is a mechanism to receive notifications
// about collisions between a body and a prop in the scene.
type SingleBodyCollisionCallback func(body Body, prop Prop, active bool)

// SingleBodyCollisionSubscription represents a notification subscription
// for single body collisions.
type SingleBodyCollisionSubscription = observer.Subscription[SingleBodyCollisionCallback]

// SingleBodyCollisionSubscriptionSet represents a set of single body
// collision subscriptions.
type SingleBodyCollisionSubscriptionSet = observer.SubscriptionSet[SingleBodyCollisionCallback]

// NewSingleBodyCollisionSubscriptionSet creates a new
// SingleBodyCollisionSubscriptionSet.
func NewSingleBodyCollisionSubscriptionSet() *SingleBodyCollisionSubscriptionSet {
	return observer.NewSubscriptionSet[SingleBodyCollisionCallback]()
}
