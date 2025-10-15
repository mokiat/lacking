package shape2d

import "github.com/mokiat/gomath/dprec"

// TODO: Write test for this!

// CheckSegmentCircleIntersection checks if the specified segment intersects
// the specified circle and returns the intersection point.
//
// This implementation assumes that the circle has backface culling. Hence, a
// segment starting from inside the circle will not produce an intersection.
// For such cases you can flip the segment to get the intersection going out.
func CheckSegmentCircleIntersection(segment Segment, circle Circle) (Intersection, bool) {
	// Solving using parametrization of the segment, resulting in a quadratic
	// equation.
	delta := dprec.Vec2Diff(segment.B, segment.A)
	offset := dprec.Vec2Diff(segment.A, circle.Position)

	// Using SqrLength in place of dot product with self.
	a := delta.SqrLength()
	b := 2.0 * dprec.Vec2Dot(delta, offset)
	c := offset.SqrLength() - dprec.Sqr(circle.Radius)

	discriminant := dprec.Sqr(b) - 4.0*a*c
	if discriminant < 0.0 {
		return Intersection{}, false
	}

	t := (-b - dprec.Sqrt(discriminant)) / (2.0 * a)
	if t < 0.0 || t > 1.0 {
		return Intersection{}, false
	}

	intersectionPoint := dprec.Vec2Lerp(segment.A, segment.B, t)

	return Intersection{
		TargetContact: intersectionPoint,
		TargetNormal: dprec.Vec2Quot(
			dprec.Vec2Diff(intersectionPoint, circle.Position),
			circle.Radius,
		),
		Depth: dprec.Vec2Diff(intersectionPoint, segment.B).Length(),
	}, true
}
