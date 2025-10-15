package shape2d

import "github.com/mokiat/gomath/dprec"

// TODO: Write test for this!

// CheckSegmentSegmentIntersection checks if a segment intersects a polygon.
func CheckSegmentSegmentIntersection(primary, secondary Segment) (Intersection, bool) {
	det := (primary.B.X-primary.A.X)*(secondary.B.Y-secondary.A.Y) -
		(primary.B.Y-primary.A.Y)*(secondary.B.X-secondary.A.X)
	if dprec.Abs(det) < 0.0001 { // parallel
		return Intersection{}, false
	}

	t := ((primary.A.Y-secondary.A.Y)*(secondary.B.X-secondary.A.X) -
		(primary.A.X-secondary.A.X)*(secondary.B.Y-secondary.A.Y)) / det
	if t < 0.0 || t > 1.0 {
		return Intersection{}, false
	}

	u := -((primary.B.X-primary.A.X)*(primary.A.Y-secondary.A.Y) -
		(primary.B.Y-primary.A.Y)*(primary.A.X-secondary.A.X)) / det
	if u < 0.0 || u > 1.0 {
		return Intersection{}, false
	}

	intersectionPoint := dprec.Vec2Lerp(primary.A, primary.B, t)
	normal := dprec.NormalVec2(dprec.Vec2Diff(secondary.B, secondary.A))

	return Intersection{
		TargetContact: intersectionPoint,
		TargetNormal:  normal,
		Depth:         dprec.Vec2Diff(intersectionPoint, primary.B).Length(),
	}, true
}
