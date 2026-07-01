package isec3d

import (
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckBoxMesh reports whether the box intersects the mesh through any of its
// triangles.
//
// Each triangle is tested with [CheckBoxTriangle], so the same one-sided,
// front-face-culled convention applies: the result is true as soon as the box
// intersects one triangle from the front. A per-triangle bounding-sphere test
// is used to skip triangles that are too far from the box to possibly
// intersect it.
func CheckBoxMesh(box shape3d.Box, mesh shape3d.Mesh) bool {
	boundingSphere := box.BoundingSphere()
	for _, triangle := range mesh.Triangles {
		if !CheckSphereSphere(boundingSphere, triangle.BoundingSphere()) {
			continue
		}
		if CheckBoxTriangle(box, triangle) {
			return true
		}
	}
	return false
}

// ResolveBoxMesh yields the contact for the triangle the box penetrates most
// deeply, if it intersects the mesh at all.
//
// Every triangle is resolved with [ResolveBoxTriangle], skipping triangles whose
// bounding sphere the box does not reach, and the resulting contacts are reduced
// to the one with the greatest Depth using a [shape3d.DeepestContact]. The box
// Depth is a true penetration distance, so this selects the deepest overlap. The
// reported [shape3d.Contact] follows the same convention as [ResolveBoxTriangle],
// with the box as the source and the triangle as the target. No contact is
// yielded when the box does not intersect any triangle from the front.
func ResolveBoxMesh(box shape3d.Box, mesh shape3d.Mesh, yield shape3d.ContactCallback) {
	boundingSphere := box.BoundingSphere()
	var deepestContact shape3d.DeepestContact
	for _, triangle := range mesh.Triangles {
		if !CheckSphereSphere(boundingSphere, triangle.BoundingSphere()) {
			continue
		}
		ResolveBoxTriangle(box, triangle, deepestContact.AddContact)
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}
