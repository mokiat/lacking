package shape2d

import "github.com/mokiat/gomath/dprec"

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
	return Circle{
		Position: transform.Apply(source.Position),
		Radius:   source.Radius,
	}
}

// Circle represents a 2D circle.
type Circle struct {

	// Position specifies the position of the circle.
	Position dprec.Vec2

	// Radius specifies the radius of the circle.
	Radius float64
}
