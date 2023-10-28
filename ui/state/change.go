package state

import "time"

// Change represents a state change.
type Change interface {
	Apply()
	Revert()
}

// ExtendableChange represents a change that can be extended
// (e.g. editbox write).
type ExtendableChange interface {

	// Extend returns whether the specified change was merged into the current
	// one.
	Extend(change Change) bool
}

// Action represents a single unit for work.
type Action func()

// ActionChange returns a Change that is based on the specified forward and
// reverse actions.
func ActionChange(accumDuration time.Duration, forward, reverse []Action) Change {
	return &actionChange{
		forward: forward,
		reverse: reverse,
	}
}

type actionChange struct {
	forward []Action
	reverse []Action
}

func (c *actionChange) Apply() {
	for _, action := range c.forward {
		action()
	}
}

func (c *actionChange) Revert() {
	for _, action := range c.reverse {
		action()
	}
}

// AccumActionChange is similar to ActionChange except that it is also
// extendable. The accumDuration interval is used to specify accumulation
// threshold for the change.
func AccumActionChange(forward, reverse []Action, accumDuration time.Duration) Change {
	return &accumActionChange{
		actionChange: actionChange{
			forward: forward,
			reverse: reverse,
		},
		when:          time.Now(),
		accumDuration: accumDuration,
	}
}

type accumActionChange struct {
	actionChange
	when          time.Time
	accumDuration time.Duration
}

func (c *accumActionChange) Extend(other Change) bool {
	otherChange, ok := other.(*accumActionChange)
	if !ok {
		return false
	}
	accumDuration := min(c.accumDuration, otherChange.accumDuration)
	if otherChange.when.Sub(c.when) > accumDuration {
		return false
	}
	c.forward = append(c.forward, otherChange.forward...)
	c.reverse = append(otherChange.reverse, c.reverse...)
	c.when = otherChange.when
	c.accumDuration = otherChange.accumDuration
	return true
}

// CombinedChange creates a composite change off of a number of smaller changes.
func CombinedChange(changes ...Change) Change {
	return &combinedChange{
		changes: changes,
	}
}

type combinedChange struct {
	changes []Change
}

func (c *combinedChange) Apply() {
	for i := 0; i < len(c.changes); i++ {
		c.changes[i].Apply()
	}
}

func (c *combinedChange) Revert() {
	for i := len(c.changes) - 1; i >= 0; i-- {
		c.changes[i].Revert()
	}
}
