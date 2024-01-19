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
