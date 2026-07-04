package isec2d

import "github.com/mokiat/lacking/core/spatial/shape2d"

// CheckRectangleMesh reports whether the rectangle intersects the mesh.
//
// The rectangle is tested against each edge of the mesh as in
// [CheckRectangleEdge], so the mesh's edges are front-face-culled: a rectangle
// is considered to intersect the mesh only if it intersects at least one edge
// from that edge's front side. Each edge's [shape2d.Edge.BoundingCircle] is used
// as a cheap broad-phase cull before the exact test, but the result is the same
// as testing every edge directly.
func CheckRectangleMesh(rectangle shape2d.Rectangle, mesh shape2d.Mesh) bool {
	boundingCircle := rectangle.BoundingCircle()
	for _, edge := range mesh.Edges {
		if !CheckCircleCircle(boundingCircle, edge.BoundingCircle()) {
			continue
		}
		if CheckRectangleEdge(rectangle, edge) {
			return true
		}
	}
	return false
}

// ResolveRectangleMesh yields the single deepest [shape2d.Contact] for the
// overlap of the rectangle with the mesh, if there is one.
//
// Each edge is resolved as in [ResolveRectangleEdge], with the rectangle as the
// source and the edge as the target, and the contact with the greatest Depth
// across all edges is reported. Moving the rectangle by that contact's Depth
// along its TargetNormal resolves the deepest overlap, though it may leave
// shallower overlaps with other edges unresolved. At most one contact is
// yielded; an empty mesh, or a rectangle that touches no edge from its front
// side, yields none.
func ResolveRectangleMesh(rectangle shape2d.Rectangle, mesh shape2d.Mesh, yield shape2d.ContactCallback) {
	boundingCircle := rectangle.BoundingCircle()
	var deepestContact shape2d.DeepestContact
	for _, edge := range mesh.Edges {
		if !CheckCircleCircle(boundingCircle, edge.BoundingCircle()) {
			continue
		}
		ResolveRectangleEdge(rectangle, edge, deepestContact.AddContact)
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}
