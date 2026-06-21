package shape2d

import "github.com/mokiat/gomath/dprec"

// Capsule represents a two-dimensional capsule shape, defined as the set of
// points within a given radius of a line segment.
type Capsule struct {
	// A is the start of the capsule's spine.
	A dprec.Vec2
	// B is the end of the capsule's spine.
	B dprec.Vec2
	// Radius specifies the radius around the spine.
	Radius float64
}

// Spine returns the line segment that forms the spine of the capsule.
func (c Capsule) Spine() Segment {
	return Segment{A: c.A, B: c.B}
}

// ContainsPoint returns whether the specified point lies within the capsule.
func (c Capsule) ContainsPoint(point dprec.Vec2) bool {
	ab := dprec.Vec2Diff(c.B, c.A)
	sqrLength := ab.SqrLength()

	var closest dprec.Vec2
	if sqrLength == 0.0 {
		closest = c.A
	} else {
		t := dprec.Clamp(dprec.Vec2Dot(dprec.Vec2Diff(point, c.A), ab)/sqrLength, 0.0, 1.0)
		closest = dprec.Vec2Sum(c.A, dprec.Vec2Prod(ab, t))
	}

	delta := dprec.Vec2Diff(point, closest)
	return delta.SqrLength() <= c.Radius*c.Radius
}

// BoundingCircle returns the smallest Circle that fully encompasses the capsule.
func (c Capsule) BoundingCircle() Circle {
	bounding := c.Spine().BoundingCircle()
	bounding.Radius += c.Radius
	return bounding
}
