package ecs

import "github.com/mokiat/lacking/game/ecs/internal"

// Condition is a predicate over an entity's component set. Conditions
// are constructed with [HasComponent], [LacksComponent], and
// [Conditions], and passed to [Scene.QueryEntities],
// [Scene.SubscribeEnter], [Scene.SubscribeExit], and
// [Scene.CheckEntity].
type Condition struct {
	positiveMask internal.TypeMask
	negativeMask internal.TypeMask
}

// HasComponent returns a condition satisfied only by entities that
// possess a component of type T.
func HasComponent[T any](compType ComponentType[T]) Condition {
	return Condition{
		positiveMask: internal.TypeMaskFromType(compType.id),
		negativeMask: internal.EmptyTypeMask(),
	}
}

// LacksComponent returns a condition satisfied only by entities that
// do not possess a component of type T.
func LacksComponent[T any](compType ComponentType[T]) Condition {
	return Condition{
		positiveMask: internal.EmptyTypeMask(),
		negativeMask: internal.TypeMaskFromType(compType.id),
	}
}

// Conditions combines multiple conditions into one that requires all
// of them to be satisfied simultaneously.
//
// Panics if the resulting condition is contradictory (e.g., both
// [HasComponent] and [LacksComponent] for the same component type).
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

// Exclusive returns a derived condition that additionally requires no
// components other than the positively-required ones to be present.
//
// This is meaningful only for conditions built with [HasComponent].
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
