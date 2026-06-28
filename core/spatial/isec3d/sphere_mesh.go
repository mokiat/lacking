package isec3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// CheckSphereMesh reports whether the sphere intersects the mesh through any of
// its triangles.
//
// Each triangle is tested with [CheckSphereTriangle], so the same one-sided,
// front-face-culled convention applies: the result is true as soon as the
// sphere intersects one triangle from the front. A per-triangle bounding-sphere
// test is used to skip triangles that are too far from the sphere to possibly
// intersect it.
func CheckSphereMesh(sphere shape3d.Sphere, mesh shape3d.Mesh) bool {
	for _, triangle := range mesh.Triangles {
		if !CheckSphereSphere(sphere, triangle.BoundingSphere()) {
			continue
		}
		if CheckSphereTriangle(sphere, triangle) {
			return true
		}
	}
	return false
}

// ResolveSphereMesh yields the contact for the triangle the sphere penetrates
// most deeply, if it intersects the mesh at all.
//
// Every triangle is resolved with [ResolveSphereTriangle], skipping triangles
// whose bounding sphere the sphere does not reach, and the resulting contacts
// are reduced to the one with the greatest Depth using a [shape3d.DeepestContact].
// Unlike the segment resolves, the sphere Depth is a true penetration distance,
// so this selects the deepest overlap. The reported [shape3d.Contact] follows
// the same convention as [ResolveSphereTriangle], with the sphere as the source
// and the triangle as the target. No contact is yielded when the sphere does
// not intersect any triangle from the front.
func ResolveSphereMesh(sphere shape3d.Sphere, mesh shape3d.Mesh, yield shape3d.ContactCallback) {
	var deepestContact shape3d.DeepestContact
	for _, triangle := range mesh.Triangles {
		if !CheckSphereSphere(sphere, triangle.BoundingSphere()) {
			continue
		}
		ResolveSphereTriangle(sphere, triangle, deepestContact.AddContact)
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}
