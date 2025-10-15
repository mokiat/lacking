package shape2d

import "github.com/mokiat/gomath/dprec"

// NewSegment creates a new segment with the specified start and end points.
func NewSegment(a, b dprec.Vec2) Segment {
	return Segment{
		A: a,
		B: b,
	}
}

// Segment represents a line segment with fixed start and end points.
type Segment struct {

	// A is the start of the segment.
	A dprec.Vec2

	// B is the end of the segment.
	B dprec.Vec2
}
