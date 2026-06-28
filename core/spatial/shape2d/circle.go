package shape2d

import "github.com/mokiat/gomath/dprec"

// Circle represents a two-dimensional circle shape.
type Circle struct {
	// Center specifies the center point of the circle.
	Center dprec.Vec2
	// Radius specifies the radius of the circle.
	Radius float64
}

// TransformedCircle returns a new [Circle] that is the result of applying the
// specified transform to the given circle. The center is moved by the transform
// while the radius is left unchanged, since a rigid-body transform preserves
// distances.
func TransformedCircle(circle Circle, transform Transform) Circle {
	return Circle{
		Center: transform.Apply(circle.Center),
		Radius: circle.Radius,
	}
}

// ContainsPoint returns whether the specified point lies within the circle.
func (c Circle) ContainsPoint(point dprec.Vec2) bool {
	delta := dprec.Vec2Diff(point, c.Center)
	return delta.SqrLength() <= c.Radius*c.Radius
}
