package ecs

import "github.com/mokiat/lacking/game/ecs/v5/internal"

// Condition represents a query condition.
type Condition struct {
	positiveMask internal.TypeMask
	negativeMask internal.TypeMask
}

// HasComponent returns a condition that requires an entity to have
// a certain component.
func HasComponent[T any](compType ComponentType[T]) Condition {
	return Condition{
		positiveMask: internal.TypeMaskFromType(compType.id),
		negativeMask: internal.EmptyTypeMask(),
	}
}

// LacksComponent returns a query condition that requires an entity to not
// have a certain component.
func LacksComponent[T any](compType ComponentType[T]) Condition {
	return Condition{
		positiveMask: internal.EmptyTypeMask(),
		negativeMask: internal.TypeMaskFromType(compType.id),
	}
}

// Conditions combines multiple conditions into a single condition that
// requires all of the individual conditions to be satisfied.
//
// It does not support having contradictory conditions (e.g. HasComponent
// and LacksComponent for the same component type).
func Conditions(conditions ...Condition) Condition {
	var result Condition
	for _, condition := range conditions {
		result.combine(condition)
	}
	if result.positiveMask.Intersects(result.negativeMask) {
		panic("contradictory conditions")
	}
	return result
}

// Exclusive returns a condition requiring that no other components are
// present other than the ones specified already.
//
// This is supported only for positive conditions (i.e. using HasComponent).
func (c Condition) Exclusive() Condition {
	c.negativeMask = c.positiveMask.Inverted()
	return c
}

func (c *Condition) combine(other Condition) {
	c.positiveMask.Combine(other.positiveMask)
	c.negativeMask.Combine(other.negativeMask)
}

func (c *Condition) isSatisfiedBy(mask internal.TypeMask) bool {
	return mask.Contains(c.positiveMask) && !mask.Intersects(c.negativeMask)
}
