package shape2d

import "github.com/mokiat/gomath/dprec"

// Circle represents a 2D circle.
type Circle struct {

	// Position specifies the position of the circle.
	Position dprec.Vec2

	// Radius specifies the radius of the circle.
	Radius float64
}

// NewCircle returns a circle from the specified position and radius.
func NewCircle(position dprec.Vec2, radius float64) Circle {
	return Circle{
		Position: position,
		Radius:   radius,
	}
}

// TransformedCircle creates a new circle based off of an existing one
// by applying the specified transformation.
func TransformedCircle(source Circle, transform Transform) Circle {
	basisTransform := transform.Basis()
	return BasisTransformedCircle(source, basisTransform)
}

// BasisTransformedCircle creates a new circle based off of an existing one
// by applying the specified basis transformation.
func BasisTransformedCircle(source Circle, transform BasisTransform) Circle {
	return Circle{
		Position: transform.Apply(source.Position),
		Radius:   source.Radius,
	}
}

// ContainsPoint checks if the circle contains the specified point.
func (c Circle) ContainsPoint(point dprec.Vec2) bool {
	delta := dprec.Vec2Diff(point, c.Position)
	return delta.SqrLength() <= c.Radius*c.Radius
}
