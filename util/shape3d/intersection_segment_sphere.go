package shape3d

import "github.com/mokiat/gomath/dprec"

// IsSegmentSphereIntersection checks if the specified segment intersects
// the specified sphere.
//
// This implementation assumes that the sphere has backface culling. Hence, a
// segment starting from inside the sphere will not produce an intersection.
// For such cases you can flip the segment to get the intersection going out.
func IsSegmentSphereIntersection(segment Segment, sphere Sphere) bool {
	// Solving using parametrization of the segment, resulting in a quadratic
	// equation.
	delta := dprec.Vec3Diff(segment.B, segment.A)
	offset := dprec.Vec3Diff(segment.A, sphere.Position)

	// Using SqrLength in place of dot product with self.
	a := delta.SqrLength()
	b := 2.0 * dprec.Vec3Dot(delta, offset)
	c := offset.SqrLength() - dprec.Sqr(sphere.Radius)

	discriminant := dprec.Sqr(b) - 4.0*a*c
	if discriminant < 0.0 {
		return false
	}

	t := (-b - dprec.Sqrt(discriminant)) / (2.0 * a)
	return t >= 0.0 && t <= 1.0
}

// CheckSegmentSphereIntersection checks if the specified segment intersects
// the specified sphere and returns the intersection point.
//
// This implementation assumes that the sphere has backface culling. Hence, a
// segment starting from inside the sphere will not produce an intersection.
// For such cases you can flip the segment to get the intersection going out.
//
// A standard Intersection result is not meaningful here.
func CheckSegmentSphereIntersection(segment Segment, sphere Sphere) (dprec.Vec3, bool) {
	// Solving using parametrization of the segment, resulting in a quadratic
	// equation.
	delta := dprec.Vec3Diff(segment.B, segment.A)
	offset := dprec.Vec3Diff(segment.A, sphere.Position)

	// Using SqrLength in place of dot product with self.
	a := delta.SqrLength()
	b := 2.0 * dprec.Vec3Dot(delta, offset)
	c := offset.SqrLength() - dprec.Sqr(sphere.Radius)

	discriminant := dprec.Sqr(b) - 4.0*a*c
	if discriminant < 0.0 {
		return dprec.Vec3{}, false
	}

	t := (-b - dprec.Sqrt(discriminant)) / (2.0 * a)
	if t < 0.0 || t > 1.0 {
		return dprec.Vec3{}, false
	}

	return dprec.Vec3Lerp(segment.A, segment.B, t), true
}
