package shape2d

import "github.com/mokiat/gomath/sprec"

// Circle represents a two-dimensional circle shape.
type Circle struct {
	// Center specifies the center point of the circle.
	Center sprec.Vec2
	// Radius specifies the radius of the circle.
	Radius float32
}

// ContainsPoint returns whether the specified point lies within the circle.
func (c Circle) ContainsPoint(point sprec.Vec2) bool {
	delta := sprec.Vec2Diff(point, c.Center)
	return delta.SqrLength() <= c.Radius*c.Radius
}
