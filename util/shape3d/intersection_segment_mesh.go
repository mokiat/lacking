package shape3d

// CheckSegmentMeshIntersection checks if a segment intersects a mesh.
//
// TODO: Rework this. Just like with sphere an Intersection is not meaningful
// here and instead an intersection point (one closest to segment start) should
// be returned.
func CheckSegmentMeshIntersection(segment Segment, mesh Mesh, yield IntersectionYieldFunc) {
	for _, triangle := range mesh.Triangles {
		CheckSegmentTriangleIntersection(segment, triangle, yield)
	}
}
