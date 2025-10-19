package shape2d

import "github.com/mokiat/gomath/dprec"

// NewSegment creates a new segment with the specified start and end points.
func NewSegment(a, b dprec.Vec2) Segment {
	return Segment{
		A: a,
		B: b,
	}
}

// TransformedSegment creates a new segment from the specified source
// segment by applying the specified transformation.
func TransformedSegment(source Segment, transform Transform) Segment {
	return Segment{
		A: transform.Apply(source.A),
		B: transform.Apply(source.B),
	}
}

// Segment represents a line segment with fixed start and end points.
type Segment struct {

	// A is the start of the segment.
	A dprec.Vec2

	// B is the end of the segment.
	B dprec.Vec2
}

// Length returns the length of the segment.
func (s *Segment) Length() float64 {
	return dprec.Vec2Diff(s.B, s.A).Length()
}

// Center returns the center point of the segment.
func (s *Segment) Center() dprec.Vec2 {
	return dprec.Vec2Prod(dprec.Vec2Sum(s.A, s.B), 0.5)
}

// BoundingCircle returns a Circle that encompasses this segment.
func (s *Segment) BoundingCircle() Circle {
	center := s.Center()
	length := s.Length()
	return NewCircle(center, length*0.5)
}
