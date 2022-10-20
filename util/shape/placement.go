package shape

// NewPlacement creates a new Placement.
func NewPlacement[T Shape](transform Transform, shape T) Placement[T] {
	return Placement[T]{
		Transform: transform,
		shape:     shape,
	}
}

// Placement represents a mechanism through which a shape can have an outside
// Transform applied. This is more performant than using TransformedShape
// which causes allocations.
type Placement[T Shape] struct {
	Transform
	shape T
}

// Shape returns the Shape held by this Placement.
func (p Placement[T]) Shape() T {
	return p.shape
}

// Transformed returns a new Placement that is based on this one but has
// the specified transform applied to it.
func (p Placement[T]) Transformed(parent Transform) Placement[T] {
	p.Transform = p.Transform.Transformed(parent)
	return p
}
