package shape3d

func CheckSphereMeshIntersection(sphere Sphere, mesh Mesh, yield IntersectionYieldFunc) {
	var collection LargestIntersection
	for _, triangle := range mesh.Triangles {
		if !IsSphereSphereIntersection(sphere, triangle.BoundingSphere()) {
			continue
		}
		CheckSphereTriangleIntersection(sphere, triangle, collection.AddIntersection)
	}
	if intersection, ok := collection.Intersection(); ok {
		yield(intersection)
	}
}
