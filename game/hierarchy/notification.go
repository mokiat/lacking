package hierarchy

import "github.com/mokiat/lacking/util/observer"

// SourceCallback is a function type for callbacks that are invoked when a
// source needs to be applied to a node.
type SourceCallback func(scene *Scene, id NodeID)

// SourceSubscription represents a notification subscription for source updates.
type SourceSubscription = observer.Subscription[SourceCallback]

// SourceSubscriptionSet represents a set of source subscriptions.
type SourceSubscriptionSet = observer.SubscriptionSet[SourceCallback]

// NewSourceSubscriptionSet creates a new SourceSubscriptionSet.
func NewSourceSubscriptionSet() *SourceSubscriptionSet {
	return observer.NewSubscriptionSet[SourceCallback]()
}

// TargetCallback is a function type for callbacks that are invoked when a
// target needs to be applied from a node.
type TargetCallback func(scene *Scene, id NodeID)

// TargetSubscription represents a notification subscription for target updates.
type TargetSubscription = observer.Subscription[TargetCallback]

// TargetSubscriptionSet represents a set of target subscriptions.
type TargetSubscriptionSet = observer.SubscriptionSet[TargetCallback]

// NewTargetSubscriptionSet creates a new TargetSubscriptionSet.
func NewTargetSubscriptionSet() *TargetSubscriptionSet {
	return observer.NewSubscriptionSet[TargetCallback]()
}

// InterpolationCallback is a function type for callbacks that are invoked when
// a target needs to be interpolated from the node.
type InterpolationCallback func(scene *Scene, id NodeID, fraction float64)

// InterpolationSubscription represents a notification subscription for
// interpolation updates.
type InterpolationSubscription = observer.Subscription[InterpolationCallback]

// InterpolationSubscriptionSet represents a set of interpolation subscriptions.
type InterpolationSubscriptionSet = observer.SubscriptionSet[InterpolationCallback]

// NewInterpolationSubscriptionSet creates a new InterpolationSubscriptionSet.
func NewInterpolationSubscriptionSet() *InterpolationSubscriptionSet {
	return observer.NewSubscriptionSet[InterpolationCallback]()
}

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
