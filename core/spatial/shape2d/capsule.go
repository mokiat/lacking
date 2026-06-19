package shape2d

import "github.com/mokiat/gomath/sprec"

// Capsule represents a two-dimensional capsule shape, defined as the set of
// points within a given radius of a line segment.
type Capsule struct {
	// A is the start of the capsule's spine.
	A sprec.Vec2
	// B is the end of the capsule's spine.
	B sprec.Vec2
	// Radius specifies the radius around the spine.
	Radius float32
}

// Spine returns the line segment that forms the spine of the capsule.
func (c Capsule) Spine() Segment {
	return Segment{A: c.A, B: c.B}
}

// ContainsPoint returns whether the specified point lies within the capsule.
func (c Capsule) ContainsPoint(point sprec.Vec2) bool {
	ab := sprec.Vec2Diff(c.B, c.A)
	sqrLength := ab.SqrLength()

	var closest sprec.Vec2
	if sqrLength == 0.0 {
		closest = c.A
	} else {
		t := sprec.Clamp(sprec.Vec2Dot(sprec.Vec2Diff(point, c.A), ab)/sqrLength, 0.0, 1.0)
		closest = sprec.Vec2Sum(c.A, sprec.Vec2Prod(ab, t))
	}

	delta := sprec.Vec2Diff(point, closest)
	return delta.SqrLength() <= c.Radius*c.Radius
}

// BoundingCircle returns the smallest Circle that fully encompasses the capsule.
func (c Capsule) BoundingCircle() Circle {
	bounding := c.Spine().BoundingCircle()
	bounding.Radius += c.Radius
	return bounding
}
