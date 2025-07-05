package shape3d

import "github.com/mokiat/gomath/dprec"

// IsSegmentCircleIntersection returns whether the specified segment intersects
// the specified circle.
//
// This implementation assumes that the circle has backface culling. Hence, a
// segment starting on the "negative" side of the circle and ending on the
// "positive" side will not produce an intersection.
func IsSegmentCircleIntersection(segment Segment, circle Circle) bool {
	surface := NewSurface(circle.Position, circle.Normal)
	point, ok := CheckSegmentSurfaceIntersection(segment, surface)
	if !ok {
		return false
	}
	distanceSqr := dprec.Vec3Diff(point, circle.Position).SqrLength()
	return distanceSqr <= dprec.Sqr(circle.Radius)
}

// CheckSegmentCircleIntersection returns whether the specified segment intersects
// the specified circle and returns the intersection point.
//
// This implementation assumes that the circle has backface culling. Hence, a
// segment starting on the "negative" side of the circle and ending on the
// "positive" side will not produce an intersection.
//
// A standard Intersection result is not meaningful here.
func CheckSegmentCircleIntersection(segment Segment, circle Circle) (dprec.Vec3, bool) {
	surface := NewSurface(circle.Position, circle.Normal)
	point, ok := CheckSegmentSurfaceIntersection(segment, surface)
	if !ok {
		return dprec.Vec3{}, false
	}
	distanceSqr := dprec.Vec3Diff(point, circle.Position).SqrLength()
	if distanceSqr > dprec.Sqr(circle.Radius) {
		return dprec.Vec3{}, false
	}
	return point, true
}
