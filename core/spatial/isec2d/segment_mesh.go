package isec2d

import "github.com/mokiat/lacking/core/spatial/shape2d"

// CheckSegmentMesh reports whether the directed segment crosses the mesh.
//
// The segment is tested against each edge of the mesh as in [CheckSegmentEdge],
// so the mesh's edges are oriented and front-face-culled: the segment is
// considered to cross the mesh only if it crosses at least one edge from that
// edge's front side, within the segment's own extent.
func CheckSegmentMesh(segment shape2d.Segment, mesh shape2d.Mesh) bool {
	for _, edge := range mesh.Edges {
		if CheckSegmentEdge(segment, edge) {
			return true
		}
	}
	return false
}

// ResolveSegmentMesh yields the single contact for the earliest point at which
// the directed segment crosses the mesh, if it crosses it at all.
//
// Each edge is resolved as in [ResolveSegmentEdge], following the entry-point
// convention where Depth is the fraction of the segment lying beyond the
// crossing. Because a greater Depth means a crossing closer to the segment's
// start, retaining the [shape2d.DeepestContact] across all edges reports the
// crossing nearest to A, that is the first edge the segment meets as it travels
// from A to B. At most one contact is yielded; an empty mesh, or a segment that
// crosses no edge from its front side, yields none.
func ResolveSegmentMesh(segment shape2d.Segment, mesh shape2d.Mesh, yield shape2d.ContactCallback) {
	var deepestContact shape2d.DeepestContact
	for _, edge := range mesh.Edges {
		ResolveSegmentEdge(segment, edge, deepestContact.AddContact)
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}
