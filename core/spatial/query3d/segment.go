package query3d

import "github.com/mokiat/gomath/sprec"

// Segment represents a line segment in 3D space that can be used for spatial
// queries.
type Segment struct {
	a sprec.Vec3
	b sprec.Vec3
}

// NewSegment creates a new Segment with the given endpoints.
func NewSegment(a, b sprec.Vec3) Segment {
	return Segment{
		a: a,
		b: b,
	}
}
