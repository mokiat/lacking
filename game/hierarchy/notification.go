package hierarchy

import "github.com/mokiat/lacking/util/observer"

// DeleteCallback is a function type for callbacks that are invoked when a
// node is about to be deleted.
type DeleteCallback func(scene *Scene, id NodeID)

// DeleteSubscription represents a notification subscription for node deletions.
type DeleteSubscription = observer.Subscription[DeleteCallback]

// DeleteSubscriptionSet represents a set of deletion subscriptions.
type DeleteSubscriptionSet = observer.SubscriptionSet[DeleteCallback]

// NewDeleteSubscriptionSet creates a new DeleteSubscriptionSet.
func NewDeleteSubscriptionSet() *DeleteSubscriptionSet {
	return observer.NewSubscriptionSet[DeleteCallback]()
}
