package shape3d

import "github.com/mokiat/gomath/dprec"

// IsSegmentSurfaceIntersection checks whether the specified segment intersects
// the specified surface.
//
// This implementation assumes that the surface has backface culling. Hence, a
// segment starting on the "negative" side of the surface and ending on the
// "positive" side will not produce an intersection.
func IsSegmentSurfaceIntersection(segment Segment, surface Surface) bool {
	point := surface.Point()

	deltaA := dprec.Vec3Diff(segment.A, point)
	heightA := dprec.Vec3Dot(deltaA, surface.Normal)
	if heightA < 0.0 {
		return false
	}

	deltaB := dprec.Vec3Diff(segment.B, point)
	heightB := -dprec.Vec3Dot(deltaB, surface.Normal)
	if heightB < 0.0 {
		return false
	}

	return heightA+heightB >= millimeter
}

// CheckSegmentSurfaceIntersection checks whether the specified segment intersects
// the specified surface and returns the intersection point.
//
// This implementation assumes that the surface has backface culling. Hence, a
// segment starting on the "negative" side of the surface and ending on the
// "positive" side will not produce an intersection.
func CheckSegmentSurfaceIntersection(segment Segment, surface Surface) (Intersection, bool) {
	point := surface.Point()

	deltaA := dprec.Vec3Diff(segment.A, point)
	heightA := dprec.Vec3Dot(deltaA, surface.Normal)
	if heightA < 0.0 {
		return Intersection{}, false
	}

	deltaB := dprec.Vec3Diff(segment.B, point)
	heightB := -dprec.Vec3Dot(deltaB, surface.Normal)
	if heightB < 0.0 {
		return Intersection{}, false
	}

	totalHeight := heightA + heightB
	if totalHeight < millimeter {
		return Intersection{}, false
	}

	intersectionPoint := dprec.Vec3Lerp(segment.A, segment.B, heightA/totalHeight)

	return Intersection{
		TargetContact: intersectionPoint,
		TargetNormal:  surface.Normal,
		Depth:         dprec.Vec3Diff(intersectionPoint, segment.B).Length(),
	}, true
}
