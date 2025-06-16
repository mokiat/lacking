package ecs5

import (
	"iter"
	"slices"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/util/mem"
)

// HasComponent returns a query condition that requires an entity to have
// a certain component.
func HasComponent[T any](set ComponentSet[T]) Condition {
	return Condition{
		positiveMask:      set.getMask(),
		negativeMask:      componentMask(0),
		isPendingDeletion: opt.Unspecified[bool](),
	}
}

// LacksComponent returns a query condition that requires an entity to not
// have a certain component.
func LacksComponent[T any](set ComponentSet[T]) Condition {
	return Condition{
		positiveMask:      componentMask(0),
		negativeMask:      set.getMask(),
		isPendingDeletion: opt.Unspecified[bool](),
	}
}

// IsHealthy returns a query condition that requires an entity to not be
// pending deletion.
func IsHealthy() Condition {
	return Condition{
		positiveMask:      componentMask(0),
		negativeMask:      componentMask(0),
		isPendingDeletion: opt.V(false),
	}
}

// IsPendingDeletion returns a query condition that requires an entity to
// have been marked for deletion.
func IsPendingDeletion() Condition {
	return Condition{
		positiveMask:      componentMask(0),
		negativeMask:      componentMask(0),
		isPendingDeletion: opt.V(true),
	}
}

// Condition represents a query condition that needs to be satisfied
// for an entity to be returned.
type Condition struct {
	positiveMask      componentMask
	negativeMask      componentMask
	isPendingDeletion opt.T[bool]
}

func (c *Condition) apply(other Condition) {
	c.positiveMask |= other.positiveMask
	c.negativeMask |= other.negativeMask
	if other.isPendingDeletion.Specified {
		c.isPendingDeletion = other.isPendingDeletion
	}
}

func (c *Condition) isSatisfied(handle *entityHandle) bool {
	if (handle.components & c.positiveMask) != c.positiveMask {
		return false
	}
	if (handle.components & c.negativeMask) != 0 {
		return false
	}
	if c.isPendingDeletion.Specified && (c.isPendingDeletion.Value != handle.isPendingDeletion) {
		return false
	}
	return true
}

// Result represents the outcome of a query operation.
//
// Make sure to call Release once you are done with it so that
// it can be reused in future searches.
type Result struct {
	allocator *mem.SparseAllocator[Result]

	// TODO: Maybe use a bitmask and scene ref instead. It will
	// take less memory but will be slower, so benchmark this first.
	entities []Entity
}

// Each invokes the callback function for each entity in this result set.
//
// While less elegant than Iter, it does not incur unnecessary allocations.
func (r *Result) Each(cb func(Entity)) {
	for _, entity := range r.entities {
		cb(entity)
	}
}

// Iter returns an iterator over the entities in this result set.
func (r *Result) Iter() iter.Seq[Entity] {
	return slices.Values(r.entities)
}

// Release frees resources allocated for this result.
func (r *Result) Release() {
	r.allocator.Release(r)
	r.allocator = nil
}
