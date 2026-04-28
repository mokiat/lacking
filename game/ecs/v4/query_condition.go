package ecs

// Condition represents a query condition.
type Condition struct {
	positiveMask componentMask
	negativeMask componentMask
}

// HasComponent returns a condition that requires an entity to have
// a certain component.
func HasComponent[T any](scene *Scene) Condition {
	tIndex := getTypeIndex[T](scene)
	return Condition{
		positiveMask: componentMaskFromType(tIndex),
		negativeMask: emptyComponentMask(),
	}
}

// LacksComponent returns a query condition that requires an entity to not
// have a certain component.
func LacksComponent[T any](scene *Scene) Condition {
	tIndex := getTypeIndex[T](scene)
	return Condition{
		positiveMask: emptyComponentMask(),
		negativeMask: componentMaskFromType(tIndex),
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

func (c *Condition) merge(other Condition) {
	c.positiveMask.addMask(other.positiveMask)
	c.negativeMask.addMask(other.negativeMask)
}

func (c *Condition) isSatisfiedBy(mask componentMask) bool {
	return mask.containsMask(c.positiveMask) &&
		!mask.intersectsMask(c.negativeMask)
}
