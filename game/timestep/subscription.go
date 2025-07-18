package timestep

import (
	"time"

	"github.com/mokiat/lacking/util/observer"
)

// UpdateCallback is a mechanism to receive delta time increments.
type UpdateCallback func(elapsedTime time.Duration)

// UpdateSubscription represents a notification subscription for updates.
type UpdateSubscription = observer.Subscription[UpdateCallback]

// UpdateSubscriptionSet represents a set of update subscriptions.
type UpdateSubscriptionSet = observer.SubscriptionSet[UpdateCallback]

// NewUpdateSubscriptionSet creates a new UpdateSubscriptionSet.
func NewUpdateSubscriptionSet() *UpdateSubscriptionSet {
	return observer.NewSubscriptionSet[UpdateCallback]()
}

// InterpolationCallback is a mechanism to receive interpolation events.
type InterpolationCallback func(fraction float64)

// InterpolationSubscription represents a notification subscription for
// interpolations.
type InterpolationSubscription = observer.Subscription[InterpolationCallback]

// InterpolationSubscriptionSet represents a set of interpolation subscriptions.
type InterpolationSubscriptionSet = observer.SubscriptionSet[InterpolationCallback]

// NewInterpolationSubscriptionSet creates a new InterpolationSubscriptionSet.
func NewInterpolationSubscriptionSet() *InterpolationSubscriptionSet {
	return observer.NewSubscriptionSet[InterpolationCallback]()
}
