package query2d

import (
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
