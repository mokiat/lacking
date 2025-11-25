package shape3d

import "github.com/mokiat/gomath/dprec"

// CheckBoxMeshIntersection checks if a Box shape intersects with a Mesh shape.
func CheckBoxMeshIntersection(box Box, mesh Mesh, yield IntersectionYieldFunc) {
	boxPosition := box.Position
	boxRotation := box.Rotation

	maxX := dprec.Vec3Prod(boxRotation.OrientationX(), box.HalfWidth)
	minX := dprec.InverseVec3(maxX)
	maxY := dprec.Vec3Prod(boxRotation.OrientationY(), box.HalfHeight)
	minY := dprec.InverseVec3(maxY)
	maxZ := dprec.Vec3Prod(boxRotation.OrientationZ(), box.HalfLength)
	minZ := dprec.InverseVec3(maxZ)

	p1 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), minZ), maxY)
	p2 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), maxZ), maxY)
	p3 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), maxZ), maxY)
	p4 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), minZ), maxY)
	p5 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), minZ), minY)
	p6 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), maxZ), minY)
	p7 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), maxZ), minY)
	p8 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), minZ), minY)

	for _, triangle := range mesh.Triangles {
		CheckSegmentTriangleIntersection(NewSegment(p1, p2), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p2, p3), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p3, p4), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p4, p1), triangle, yield)

		CheckSegmentTriangleIntersection(NewSegment(p5, p6), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p6, p7), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p7, p8), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p8, p5), triangle, yield)

		CheckSegmentTriangleIntersection(NewSegment(p1, p5), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p2, p6), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p3, p7), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p4, p8), triangle, yield)

		// since segment intersections are unidirectional, check the opposite direction as well

		CheckSegmentTriangleIntersection(NewSegment(p2, p1), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p3, p2), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p4, p3), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p1, p4), triangle, yield)

		CheckSegmentTriangleIntersection(NewSegment(p6, p5), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p7, p6), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p8, p7), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p5, p8), triangle, yield)

		CheckSegmentTriangleIntersection(NewSegment(p5, p1), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p6, p2), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p7, p3), triangle, yield)
		CheckSegmentTriangleIntersection(NewSegment(p8, p4), triangle, yield)
	}
}
