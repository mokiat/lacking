package shape2d

import "github.com/mokiat/gomath/dprec"

// Circle represents a two-dimensional circle shape.
type Circle struct {
	// Center specifies the center point of the circle.
	Center dprec.Vec2
	// Radius specifies the radius of the circle.
	Radius float64
}

// ContainsPoint returns whether the specified point lies within the circle.
func (c Circle) ContainsPoint(point dprec.Vec2) bool {
	delta := dprec.Vec2Diff(point, c.Center)
	return delta.SqrLength() <= c.Radius*c.Radius
}
