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

// IsSegmentSphereOverlap checks if the specified segment overlaps in
// any way the specified sphere, including being contained by the sphere.
func IsSegmentSphereOverlap(segment Segment, sphere Sphere) bool {
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

	t1 := (-b - dprec.Sqrt(discriminant)) / (2.0 * a)
	t2 := (-b + dprec.Sqrt(discriminant)) / (2.0 * a)
	return t1 < 1.0 && t2 > 0.0
}

// CheckSegmentSphereIntersection checks if the specified segment intersects
// the specified sphere and returns the intersection point.
//
// This implementation assumes that the sphere has backface culling. Hence, a
// segment starting from inside the sphere will not produce an intersection.
// For such cases you can flip the segment to get the intersection going out.
func CheckSegmentSphereIntersection(segment Segment, sphere Sphere) (Intersection, bool) {
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
		return Intersection{}, false
	}

	t := (-b - dprec.Sqrt(discriminant)) / (2.0 * a)
	if t < 0.0 || t > 1.0 {
		return Intersection{}, false
	}

	intersectionPoint := dprec.Vec3Lerp(segment.A, segment.B, t)

	return Intersection{
		TargetContact: intersectionPoint,
		TargetNormal: dprec.Vec3Quot(
			dprec.Vec3Diff(intersectionPoint, sphere.Position),
			sphere.Radius,
		),
		Depth: dprec.Vec3Diff(intersectionPoint, segment.B).Length(),
	}, true
}
