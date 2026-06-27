package shape2d

import "github.com/mokiat/gomath/dprec"

// Segment represents a line segment with fixed start and end points.
type Segment struct {
	// A is the start of the segment.
	A dprec.Vec2
	// B is the end of the segment.
	B dprec.Vec2
}

// TransformedSegment returns a new Segment that is the result of applying the
// specified transform to both endpoints of the given segment.
func TransformedSegment(segment Segment, transform Transform) Segment {
	return Segment{
		A: transform.Apply(segment.A),
		B: transform.Apply(segment.B),
	}
}

// Length returns the length of the segment.
func (s Segment) Length() float64 {
	return dprec.Vec2Diff(s.B, s.A).Length()
}

// Midpoint returns the point halfway between the start and end of the segment.
func (s Segment) Midpoint() dprec.Vec2 {
	return dprec.Vec2Prod(dprec.Vec2Sum(s.A, s.B), 0.5)
}

// Flipped returns a new Segment with the start and end points swapped.
func (s Segment) Flipped() Segment {
	return Segment{
		A: s.B,
		B: s.A,
	}
}

// BoundingCircle returns the smallest Circle that fully encompasses the segment.
func (s Segment) BoundingCircle() Circle {
	return Circle{
		Center: s.Midpoint(),
		Radius: s.Length() * 0.5,
	}
}
