package shape3d

import "github.com/mokiat/gomath/dprec"

// CheckSegmentTriangleIntersection checks if a line segment shape intersects
// with a triangle shape.
//
// This function assumes that the triangle has backface culling.
func CheckSegmentTriangleIntersection(segment Segment, triangle Triangle, yield IntersectionYieldFunc) {
	// Using the Möller–Trumbore intersection algorithm.

	vecAB := dprec.Vec3Diff(triangle.B, triangle.A)
	vecAC := dprec.Vec3Diff(triangle.C, triangle.A)

	dir := dprec.Vec3Diff(segment.B, segment.A)
	crossDirVecAC := dprec.Vec3Cross(dir, vecAC)

	det := dprec.Vec3Dot(vecAB, crossDirVecAC)
	if det < 0.00001 { // backface culling
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

	intersectionPoint := dprec.Vec3Sum(segment.A, dprec.Vec3Prod(dir, t/det))

	normal := dprec.ResizedVec3(dprec.Vec3Cross(vecAB, vecAC), 1.0)

	yield(Intersection{
		TargetContact: intersectionPoint,
		TargetNormal:  normal,
		Depth: dprec.Vec3Dot(
			dprec.Vec3Diff(intersectionPoint, segment.B),
			normal,
		),
	})
}
