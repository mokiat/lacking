package query2d

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
)

// Segment represents a line segment in 2D space that can be used for spatial
// queries.
type Segment struct {
	a sprec.Vec2
	b sprec.Vec2
}

// NewSegment creates a new Segment with the given endpoints.
func NewSegment(a, b sprec.Vec2) Segment {
	return Segment{
		a: a,
		b: b,
	}
}

// String returns a string representation of the Segment.
func (s Segment) String() string {
	return fmt.Sprintf("(%f, %f)", s.a, s.b)
}
