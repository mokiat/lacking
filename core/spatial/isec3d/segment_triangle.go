package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSegmentTriangle checks whether the given segment intersects the given
// triangle and returns true if it does.
//
// The segment is treated as a directed, front-face-culled probe from A to B:
// only a crossing that enters through the side the triangle's normal faces is
// considered an intersection. A segment that approaches the triangle from
// behind, or that does not reach the triangle's plane within its A-to-B span,
// is not considered to intersect it.
func CheckSegmentTriangle(segment shape3d.Segment, triangle shape3d.Triangle) bool {
	// Using the Moller-Trumbore intersection algorithm.
	vecAB := dprec.Vec3Diff(triangle.B, triangle.A)
	vecAC := dprec.Vec3Diff(triangle.C, triangle.A)

	dir := dprec.Vec3Diff(segment.B, segment.A)
	crossDirVecAC := dprec.Vec3Cross(dir, vecAC)

	det := dprec.Vec3Dot(vecAB, crossDirVecAC)
	if det <= 0.0 { // backface culling
		return false
	}

	offset := dprec.Vec3Diff(segment.A, triangle.A)

	u := dprec.Vec3Dot(offset, crossDirVecAC)
	if u < 0.0 || u > det {
		return false
	}

	crossOffsetVecAB := dprec.Vec3Cross(offset, vecAB)

	v := dprec.Vec3Dot(dir, crossOffsetVecAB)
	if v < 0.0 || (u+v) > det {
		return false
	}

	t := dprec.Vec3Dot(vecAC, crossOffsetVecAB)
	if t < 0.0 || t > det {
		return false
	}

	return true
}

// ResolveSegmentTriangle checks whether the given segment intersects the given
// triangle and, if it does, invokes yield with the resulting Contact.
//
// The segment (source) is treated as a directed, front-face-culled probe from A
// to B against the triangle (target), following the same convention as
// CheckSegmentTriangle. When a front-facing crossing is found within the segment
// span, the reported contact has its TargetPoint at the point where the segment
// crosses the triangle's plane and its TargetNormal equal to the triangle's
// outward (front-facing) normal. Its Depth is the fraction of the segment lying
// beyond that crossing, in the range [0, 1] (1 when the segment crosses at A, 0
// when it crosses at B); being a fraction, it is comparable across shapes, so
// DeepestContact selects the earliest crossing along the segment. No contact is
// yielded otherwise.
func ResolveSegmentTriangle(segment shape3d.Segment, triangle shape3d.Triangle, yield shape3d.ContactCallback) {
	// Using the Moller-Trumbore intersection algorithm.
	vecAB := dprec.Vec3Diff(triangle.B, triangle.A)
	vecAC := dprec.Vec3Diff(triangle.C, triangle.A)

	dir := dprec.Vec3Diff(segment.B, segment.A)
	crossDirVecAC := dprec.Vec3Cross(dir, vecAC)

	det := dprec.Vec3Dot(vecAB, crossDirVecAC)
	if det <= 0.0 { // backface culling
		return
	}

	offset := dprec.Vec3Diff(segment.A, triangle.A)

	u := dprec.Vec3Dot(offset, crossDirVecAC)
	if u < 0.0 || u > det {
		return
	}

	crossOffsetVecAB := dprec.Vec3Cross(offset, vecAB)

	v := dprec.Vec3Dot(dir, crossOffsetVecAB)
	if v < 0.0 || (u+v) > det {
		return
	}

	t := dprec.Vec3Dot(vecAC, crossOffsetVecAB)
	if t < 0.0 || t > det {
		return
	}
	tNormalized := t / det

	contactPoint := dprec.Vec3Sum(segment.A, dprec.Vec3Prod(dir, tNormalized))
	normal := dprec.ResizedVec3(dprec.Vec3Cross(vecAB, vecAC), 1.0)

	yield(shape3d.Contact{
		TargetPoint:  contactPoint,
		TargetNormal: normal,
		Depth:        1.0 - tNormalized,
	})
}
