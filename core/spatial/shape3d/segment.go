package shape3d

import "github.com/mokiat/gomath/dprec"

// Segment represents a line segment with fixed start and end points.
type Segment struct {
	// A is the start of the segment.
	A dprec.Vec3
	// B is the end of the segment.
	B dprec.Vec3
}

// Length returns the length of the segment.
func (s Segment) Length() float64 {
	return dprec.Vec3Diff(s.B, s.A).Length()
}

// Midpoint returns the point halfway between the start and end of the segment.
func (s Segment) Midpoint() dprec.Vec3 {
	return dprec.Vec3Prod(dprec.Vec3Sum(s.A, s.B), 0.5)
}

// BoundingSphere returns the smallest Sphere that fully encompasses the segment.
func (s Segment) BoundingSphere() Sphere {
	return Sphere{
		Center: s.Midpoint(),
		Radius: s.Length() * 0.5,
	}
}
