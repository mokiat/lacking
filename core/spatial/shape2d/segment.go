package shape2d

import "github.com/mokiat/gomath/sprec"

// Segment represents a line segment with fixed start and end points.
type Segment struct {
	// A is the start of the segment.
	A sprec.Vec2
	// B is the end of the segment.
	B sprec.Vec2
}

// Length returns the length of the segment.
func (s Segment) Length() float32 {
	return sprec.Vec2Diff(s.B, s.A).Length()
}

// Midpoint returns the point halfway between the start and end of the segment.
func (s Segment) Midpoint() sprec.Vec2 {
	return sprec.Vec2Prod(sprec.Vec2Sum(s.A, s.B), 0.5)
}

// BoundingCircle returns the smallest Circle that fully encompasses the segment.
func (s Segment) BoundingCircle() Circle {
	return Circle{
		Center: s.Midpoint(),
		Radius: s.Length() * 0.5,
	}
}
