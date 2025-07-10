package shape3d

// CheckSegmentMeshIntersection checks if a segment intersects a mesh.
//
// TODO: Rework this. Just like with sphere an Intersection is not meaningful
// here and instead an intersection point (one closest to segment start) should
// be returned.
func CheckSegmentMeshIntersection(segment Segment, mesh Mesh) (Intersection, bool) {
	var collection WorstIntersection
	for _, triangle := range mesh.Triangles {
		if intersection, ok := CheckSegmentTriangleIntersection(segment, triangle); ok {
			collection.AddIntersection(intersection)
		}
	}
	return collection.Intersection()
}
