package shape3d

func CheckSphereMeshIntersection(sphere Sphere, mesh Mesh) (Intersection, bool) {
	var collection LargestIntersection
	for _, triangle := range mesh.Triangles {
		if !IsSphereSphereIntersection(sphere, triangle.BoundingSphere()) {
			continue
		}
		if intersection, ok := CheckSphereTriangleIntersection(sphere, triangle); ok {
			collection.AddIntersection(intersection)
		}
	}
	return collection.Intersection()
}
