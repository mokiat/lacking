package mvc

import (
	"fmt"

	"github.com/mokiat/lacking/util/filter"
)

// Change represents a notification that something has changed.
type Change interface {

	// Description returns a string description of what has changed.
	Description() string
}

// IsChange checks whether the specified Change or any of its parents
// is equal to the target Change.
func IsChange(change, target Change) bool {
	if change == target {
		return true
	}
	comparable, ok := change.(interface {
		Is(Change) bool
	})
	if ok && comparable.Is(target) {
		return true
	}
	parentable, ok := change.(interface {
		Parent() Change
	})
	if !ok {
		return false
	}
	parent := parentable.Parent()
	if parent == nil {
		return false
	}
	return IsChange(parent, target)
}

// NewChange creates a new Change instance with the specified description.
func NewChange(description string) Change {
	return &stringChange{
		description: description,
	}
}

type stringChange struct {
	description string
}

func (c *stringChange) Description() string {
	return c.description
}

// SubChange creates a new Change that extends an existing Change
// and adds the specified description.
func SubChange(parent Change, description string) Change {
	return &subChange{
		parent:      parent,
		description: description,
	}
}

type subChange struct {
	parent      Change
	description string
}

func (c *subChange) Description() string {
	return fmt.Sprintf("%s: %s", c.parent.Description(), c.description)
}

func (c *subChange) Parent() Change {
	return c.parent
}

// MultiChange represents a Change that groups multiple other Changes.
type MultiChange struct {

	// Changes is a slice of Changes that make up this MultiChange.
	Changes []Change
}

func (c *MultiChange) Description() string {
	return fmt.Sprintf("multi-change (%d)", len(c.Changes))
}

func (c *MultiChange) Is(target Change) bool {
	for _, candidate := range c.Changes {
		if IsChange(candidate, target) {
			return true
		}
	}
	return false
}

// ChangeFilter is a filter for Change objects.
type ChangeFilter = filter.Func[Change]
