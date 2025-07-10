package shape3d

import "github.com/mokiat/gomath/dprec"

// CheckSegmentTriangleIntersection checks if a line segment shape intersects
// with a triangle shape.
func CheckSegmentTriangleIntersection(segment Segment, triangle Triangle) (Intersection, bool) {
	normal := triangle.Normal()
	pointA := segment.A
	pointB := segment.B

	heightA := dprec.Vec3Dot(normal, dprec.Vec3Diff(pointA, triangle.A))
	heightB := dprec.Vec3Dot(normal, dprec.Vec3Diff(pointB, triangle.A))

	if (heightA > 0.0 && heightB > 0.0) || (heightA < 0.0 && heightB < 0.0) {
		return Intersection{}, false
	}
	if heightA < 0.0 {
		pointA, pointB = pointB, pointA
		heightA, heightB = heightB, heightA
	}

	projectedPoint := dprec.Vec3Sum(
		dprec.Vec3Prod(pointA, -heightB/(heightA-heightB)),
		dprec.Vec3Prod(pointB, heightA/(heightA-heightB)),
	)

	if !triangle.ContainsPoint(projectedPoint) {
		return Intersection{}, false
	}

	return Intersection{
		TargetContact: projectedPoint,
		TargetNormal:  normal,
		Depth:         -heightB,
	}, true
}
