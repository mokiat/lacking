package isec2d

import "github.com/mokiat/lacking/core/spatial/shape2d"

// CheckCircleMesh reports whether the circle intersects the mesh.
//
// The circle is tested against each edge of the mesh as in [CheckCircleEdge],
// so the mesh's edges are front-face-culled: a circle is considered to intersect
// the mesh only if it intersects at least one edge from that edge's front side.
// Each edge's [shape2d.Edge.BoundingCircle] is used as a cheap broad-phase cull
// before the exact test, but the result is the same as testing every edge
// directly.
func CheckCircleMesh(circle shape2d.Circle, mesh shape2d.Mesh) bool {
	for _, edge := range mesh.Edges {
		if !CheckCircleCircle(circle, edge.BoundingCircle()) {
			continue
		}
		if CheckCircleEdge(circle, edge) {
			return true
		}
	}
	return false
}

// ResolveCircleMesh yields the single deepest [shape2d.Contact] for the overlap
// of the circle with the mesh, if there is one.
//
// Each edge is resolved as in [ResolveCircleEdge], with the circle as the source
// and the edge as the target, and the contact with the greatest Depth across all
// edges is reported. Moving the circle by that contact's Depth along its
// TargetNormal resolves the deepest overlap, though it may leave shallower
// overlaps with other edges unresolved. At most one contact is yielded; an empty
// mesh, or a circle that touches no edge from its front side, yields none.
func ResolveCircleMesh(circle shape2d.Circle, mesh shape2d.Mesh, yield shape2d.ContactCallback) {
	var deepestContact shape2d.DeepestContact
	for _, edge := range mesh.Edges {
		if !CheckCircleCircle(circle, edge.BoundingCircle()) {
			continue
		}
		ResolveCircleEdge(circle, edge, deepestContact.AddContact)
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}
