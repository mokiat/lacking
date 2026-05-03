package ecs

// Condition represents a query condition.
type Condition struct {
	positiveMask componentMask
	negativeMask componentMask
}

// HasComponent returns a condition that requires an entity to have
// a certain component.
func HasComponent[T any](compType *ComponentType[T]) Condition {
	id := compType.id()
	return Condition{
		positiveMask: componentMaskFromType(id),
		negativeMask: emptyComponentMask(),
	}
}

// LacksComponent returns a query condition that requires an entity to not
// have a certain component.
func LacksComponent[T any](compType *ComponentType[T]) Condition {
	id := compType.id()
	return Condition{
		positiveMask: emptyComponentMask(),
		negativeMask: componentMaskFromType(id),
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
		result.merge(condition)
	}
	if result.positiveMask.intersectsMask(result.negativeMask) {
		panic("contradictory conditions")
	}
	return result
}

// Exclusive returns a condition requiring that no other components are
// present other than the ones specified already.
//
// This is supported only for positive conditions (i.e. using HasComponent).
func (c Condition) Exclusive() Condition {
	c.negativeMask = c.positiveMask.inverted()
	return c
}

func (c *Condition) merge(other Condition) {
	c.positiveMask.addMask(other.positiveMask)
	c.negativeMask.addMask(other.negativeMask)
}

func (c *Condition) isSatisfiedBy(mask componentMask) bool {
	return mask.containsMask(c.positiveMask) && !mask.intersectsMask(c.negativeMask)
}
