package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSegmentSphere reports whether the directed segment enters the sphere.
//
// Like all segment checks in this package, the test is oriented and
// face-culled: the segment is treated as a directed probe from A to B and only
// the point where it crosses into the sphere's surface from outside, within the
// segment's extent, counts as an intersection. A segment that lies entirely
// inside the sphere, or whose start is already inside it, is not considered to
// intersect it.
func CheckSegmentSphere(segment shape3d.Segment, sphere shape3d.Sphere) bool {
	// Solving using parametrization of the segment, resulting in a quadratic
	// equation.
	delta := dprec.Vec3Diff(segment.B, segment.A)
	offset := dprec.Vec3Diff(segment.A, sphere.Center)

	// Using SqrLength in place of dot product with self.
	a := delta.SqrLength()
	if a == 0.0 {
		return false // degenerate segment
	}
	b := 2.0 * dprec.Vec3Dot(delta, offset)
	c := offset.SqrLength() - dprec.Sqr(sphere.Radius)

	discriminant := dprec.Sqr(b) - 4.0*a*c
	if discriminant < 0.0 {
		return false
	}

	t := (-b - dprec.Sqrt(discriminant)) / (2.0 * a)
	return t >= 0.0 && t <= 1.0
}

// CheckSegmentSphereOverlap reports whether the segment and the sphere overlap
// in any way.
//
// Unlike [CheckSegmentSphere], this test is neither oriented nor face-culled: it
// returns true whenever any part of the segment lies within the sphere,
// including when the segment lies entirely inside it or starts inside it and
// exits. The result does not depend on the order of the segment's endpoints.
//
// Use it when only the fact of an overlap matters; use [CheckSegmentSphere] or
// [ResolveSegmentSphere] when the directed entry point is needed.
func CheckSegmentSphereOverlap(segment shape3d.Segment, sphere shape3d.Sphere) bool {
	// Solving using parametrization of the segment, resulting in a quadratic
	// equation.
	delta := dprec.Vec3Diff(segment.B, segment.A)
	offset := dprec.Vec3Diff(segment.A, sphere.Center)

	// Using SqrLength in place of dot product with self.
	a := delta.SqrLength()
	b := 2.0 * dprec.Vec3Dot(delta, offset)
	c := offset.SqrLength() - dprec.Sqr(sphere.Radius)

	if a == 0.0 {
		return c <= 0.0 // degenerate segment; check if the point is inside the sphere
	}

	discriminant := dprec.Sqr(b) - 4.0*a*c
	if discriminant < 0.0 {
		return false
	}

	sqrtDiscriminant := dprec.Sqrt(discriminant)
	fraction := 1.0 / (2.0 * a)
	t1 := (-b - sqrtDiscriminant) * fraction
	t2 := (-b + sqrtDiscriminant) * fraction
	return t1 <= 1.0 && t2 >= 0.0
}

// ResolveSegmentSphere yields the contact at which the directed segment enters
// the sphere, if it enters it at all.
//
// The contact follows the entry-point convention shared by the segment Resolve
// routines in this package:
//
//   - TargetPoint is the point where the segment first crosses into the sphere.
//   - TargetNormal is the outward surface normal of the sphere there.
//   - Depth is the fraction of the segment lying beyond the entry point, in the
//     range [0, 1] (1 when the segment enters at A, 0 when it enters at B). It
//     is comparable across shapes, so DeepestContact selects the earliest entry
//     along the segment.
func ResolveSegmentSphere(segment shape3d.Segment, sphere shape3d.Sphere, yield shape3d.ContactCallback) {
	// Solving using parametrization of the segment, resulting in a quadratic
	// equation.
	delta := dprec.Vec3Diff(segment.B, segment.A)
	offset := dprec.Vec3Diff(segment.A, sphere.Center)

	// Using SqrLength in place of dot product with self.
	a := delta.SqrLength()
	if a == 0.0 {
		return // degenerate segment
	}
	b := 2.0 * dprec.Vec3Dot(delta, offset)
	c := offset.SqrLength() - dprec.Sqr(sphere.Radius)

	discriminant := dprec.Sqr(b) - 4.0*a*c
	if discriminant < 0.0 {
		return
	}

	t := (-b - dprec.Sqrt(discriminant)) / (2.0 * a)
	if t < 0.0 || t > 1.0 {
		return
	}

	contactPoint := dprec.Vec3Lerp(segment.A, segment.B, t)
	normal := dprec.Vec3Quot(
		dprec.Vec3Diff(contactPoint, sphere.Center),
		sphere.Radius,
	)

	yield(shape3d.Contact{
		TargetPoint:  contactPoint,
		TargetNormal: normal,
		Depth:        1.0 - t,
	})
}
