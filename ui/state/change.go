package state

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

// FuncChange is a utility function that creates a change based on the
// specified apply and revert functions.
func FuncChange(apply, revert func()) Change {
	return &funcChange{
		apply:  apply,
		revert: revert,
	}
}

type funcChange struct {
	apply  func()
	revert func()
}

func (ch *funcChange) Apply() {
	ch.apply()
}

func (ch *funcChange) Revert() {
	ch.revert()
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
