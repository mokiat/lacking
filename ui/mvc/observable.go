package mvc

import "github.com/mokiat/gog/filter"

// Observable represents something that can be observed for changes.
type Observable interface {

	// Subscribe registers the specified callback function to this Observable.
	// It will only be called if the inbound Change passes the specified filters.
	// The returned Subscription can be used to unsubscribe from this Target.
	Subscribe(callback Callback, filters ...ChangeFilter) Subscription

	// SignalChange notifies all Subscribers of the specified Change.
	SignalChange(change Change)

	// AccumulateChanges calls the specified closure function and any calls
	// to SignalChange during that invocation will not be sent but instead will
	// be recorded. Once the closure is complete then all recorded changes will
	// be sent through a single MultiChange change, as long as the closure
	// does not return an error.
	AccumulateChanges(fn func() error) error
}

// Subscription represents an association of a Callback function to an
// Observable.
type Subscription interface {

	// Delete removes this Subscription from the associated Observable and the
	// Callback of this Subscription will no longer be called.
	Delete()
}

// Callback is a function to be called when a Change is received by the
// Subscription.
type Callback func(change Change)

// NewObservable creates a new Observable instance.
func NewObservable() Observable {
	return &observable{}
}

type observable struct {
	firstSubscription  *subscription
	accumulationDepth  int
	accumulatedChanges []Change
}

func (o *observable) Subscribe(cb Callback, fltrs ...ChangeFilter) Subscription {
	subscription := &subscription{
		obs:  o,
		next: o.firstSubscription,
		fltr: filter.And(fltrs...),
		cb:   cb,
	}
	o.firstSubscription = subscription
	return subscription
}

func (o *observable) SignalChange(change Change) {
	if o.accumulationDepth > 0 {
		o.accumulatedChanges = append(o.accumulatedChanges, change)
		return
	}
	current := o.firstSubscription
	for current != nil {
		if current.fltr(change) {
			current.cb(change)
		}
		current = current.next
	}
}

func (o *observable) AccumulateChanges(fn func() error) error {
	var err error
	defer func() {
		if o.accumulationDepth == 0 {
			if err == nil {
				o.SignalChange(&MultiChange{
					Changes: o.accumulatedChanges,
				})
			}
			o.accumulatedChanges = o.accumulatedChanges[:0]
		}
	}()
	o.accumulationDepth++
	defer func() {
		o.accumulationDepth--
	}()
	err = fn()
	return err
}

func (o *observable) unsubscribe(sub *subscription) {
	if o.firstSubscription == sub {
		o.firstSubscription = sub.next
		return
	}
	current := o.firstSubscription
	for current != nil {
		if current.next == sub {
			current.next = sub.next
			return
		}
		current = current.next
	}
}

type subscription struct {
	obs  *observable
	next *subscription
	fltr ChangeFilter
	cb   Callback
}

func (s *subscription) Delete() {
	s.obs.unsubscribe(s)
}
