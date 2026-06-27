package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSegmentSphere reports whether the segment intersects the sphere.
//
// A segment that lies entirely inside the sphere, or that merely touches its
// surface, is considered to intersect it.
func CheckSegmentSphere(segment shape3d.Segment, sphere shape3d.Sphere) bool {
	// Find the point on the segment closest to the sphere center and test
	// whether it lies within the radius. This avoids the square root and the
	// quadratic solve that finding the actual intersection points would need.
	delta := dprec.Vec3Diff(segment.B, segment.A)
	offset := dprec.Vec3Diff(sphere.Center, segment.A)

	sqrLength := delta.SqrLength()
	var closest dprec.Vec3
	if sqrLength == 0 {
		closest = segment.A // degenerate, zero-length segment
	} else {
		factor := dprec.Clamp(dprec.Vec3Dot(offset, delta)/sqrLength, 0.0, 1.0)
		closest = dprec.Vec3Sum(segment.A, dprec.Vec3Prod(delta, factor))
	}
	return dprec.Vec3Diff(closest, sphere.Center).SqrLength() <= dprec.Sqr(sphere.Radius)
}

// ResolveSegmentSphere yields a Contact for the overlap of the segment with the
// sphere, if there is one.
//
// The contact is derived from the point on the segment closest to the sphere
// center, so the result is symmetric with respect to the segment's endpoints and
// is reported whenever the two overlap, consistent with CheckSegmentSphere. The
// contact is expressed relative to the sphere as the target shape: TargetPoint
// is the point on the sphere's surface closest to the segment, TargetNormal is
// the outward surface normal there, and Depth is how far the segment reaches
// past the surface along that normal.
func ResolveSegmentSphere(segment shape3d.Segment, sphere shape3d.Sphere, yield shape3d.ContactCallback) {
	delta := dprec.Vec3Diff(segment.B, segment.A)
	offset := dprec.Vec3Diff(sphere.Center, segment.A)

	sqrLength := delta.SqrLength()
	var closest dprec.Vec3
	if sqrLength == 0 {
		closest = segment.A // degenerate, zero-length segment
	} else {
		factor := dprec.Clamp(dprec.Vec3Dot(offset, delta)/sqrLength, 0.0, 1.0)
		closest = dprec.Vec3Sum(segment.A, dprec.Vec3Prod(delta, factor))
	}

	centerToClosest := dprec.Vec3Diff(closest, sphere.Center)
	distance := centerToClosest.Length()
	if distance > sphere.Radius {
		return // the segment does not reach the sphere
	}

	var normal dprec.Vec3
	switch {
	case distance > 0:
		normal = dprec.Vec3Quot(centerToClosest, distance)
	case sqrLength > 0:
		// The center lies on the segment; the separation normal is not unique.
		// Any direction perpendicular to the segment separates them, so pick a
		// deterministic one.
		normal = dprec.NormalVec3(delta)
	default:
		// Fully degenerate: a zero-length segment located at the sphere center.
		normal = dprec.BasisXVec3()
	}

	yield(shape3d.Contact{
		TargetPoint:  dprec.Vec3Sum(sphere.Center, dprec.Vec3Prod(normal, sphere.Radius)),
		TargetNormal: normal,
		Depth:        sphere.Radius - distance,
	})
}
