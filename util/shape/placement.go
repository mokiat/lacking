package shape

// NewPlacement creates a new Placement.
func NewPlacement(transform Transform, shape Shape) Placement {
	return Placement{
		shape:     shape,
		transform: transform,
	}
}

// Placement represents a mechanism through which a shape can have an outside
// Transform applied. This is more performant than using TransformedShape
// which causes allocations.
type Placement struct {
	transform Transform
	shape     Shape
}

// Transform returns the Transform of this Placement.
func (p Placement) Transform() Transform {
	return p.transform
}

// Shape returns the Shape held by this Placement.
func (p Placement) Shape() Shape {
	return p.shape
}
