package isec3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// CheckSegmentMesh reports whether the directed segment enters the mesh through
// any of its triangles.
//
// Each triangle is tested as the same directed, front-face-culled probe as
// [CheckSegmentTriangle], so the result is true as soon as the segment crosses
// one triangle from the front within its A-to-B span. A triangle reached only
// from behind, or beyond the segment's extent, does not count.
func CheckSegmentMesh(segment shape3d.Segment, mesh shape3d.Mesh) bool {
	for _, triangle := range mesh.Triangles {
		if CheckSegmentTriangle(segment, triangle) {
			return true
		}
	}
	return false
}

// ResolveSegmentMesh yields the contact where the directed segment first enters
// the mesh, if it enters it at all.
//
// Every triangle is resolved with [ResolveSegmentTriangle] and the resulting
// contacts are reduced to the one closest to the segment's start. Because the
// segment resolves report Depth as the fraction of the segment beyond the
// crossing, the earliest entry is the one with the greatest Depth, so the
// triangle contacts are gathered with a [shape3d.DeepestContact]. The reported
// [shape3d.Contact] follows the same convention as [ResolveSegmentTriangle]. No
// contact is yielded when no triangle is crossed from the front within the
// segment.
func ResolveSegmentMesh(segment shape3d.Segment, mesh shape3d.Mesh, yield shape3d.ContactCallback) {
	var deepestContact shape3d.DeepestContact
	for _, triangle := range mesh.Triangles {
		ResolveSegmentTriangle(segment, triangle, deepestContact.AddContact)
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}
