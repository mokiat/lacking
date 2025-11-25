package shape3d

import "github.com/mokiat/gomath/dprec"

// IsSegmentCircleIntersection returns whether the specified segment intersects
// the specified circle.
//
// This implementation assumes that the circle has backface culling. Hence, a
// segment starting on the "negative" side of the circle and ending on the
// "positive" side will not produce an intersection.
func IsSegmentCircleIntersection(segment Segment, circle Circle) bool {
	var collection LargestIntersection
	surface := NewSurface(circle.Position, circle.Normal)
	CheckSegmentSurfaceIntersection(segment, surface, collection.AddIntersection)
	intersection, ok := collection.Intersection()
	if !ok {
		return false
	}
	distanceSqr := dprec.Vec3Diff(intersection.TargetContact, circle.Position).SqrLength()
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
func CheckSegmentCircleIntersection(segment Segment, circle Circle, yield IntersectionYieldFunc) {
	var collection LargestIntersection
	surface := NewSurface(circle.Position, circle.Normal)
	CheckSegmentSurfaceIntersection(segment, surface, collection.AddIntersection)
	intersection, ok := collection.Intersection()
	if !ok {
		return
	}
	distanceSqr := dprec.Vec3Diff(intersection.TargetContact, circle.Position).SqrLength()
	if distanceSqr > dprec.Sqr(circle.Radius) {
		return
	}
	yield(intersection)
}
