package query3d

import "github.com/mokiat/gomath/dprec"

// Segment represents a line segment in 3D space that can be used for spatial
// queries.
type Segment struct {
	a dprec.Vec3
	b dprec.Vec3
}

// NewSegment creates a new [Segment] with the given endpoints.
func NewSegment(a, b dprec.Vec3) Segment {
	return Segment{
		a: a,
		b: b,
	}
}
