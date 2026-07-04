package internal

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// Hull is a convex polyhedron in local space together with its world-space
// rotation, used by the GJK solver to compute support points.
type Hull struct {
	// Rotation is the orientation of the hull in world space.
	Rotation shape3d.Rotation
	// InvRotation is the inverse of Rotation, cached to avoid recomputing it
	// on every support query.
	InvRotation shape3d.Rotation
	// Points holds the local-space vertices of the convex polyhedron core.
	Points []dprec.Vec3
}

// Support returns the world-space position and the index of the hull
// vertex that is furthest along dir. The direction is expected to be in
// world space and does not need to be normalized.
func (h *Hull) Support(dir dprec.Vec3) (dprec.Vec3, int) {
	dir = h.InvRotation.Apply(dir)
	bestIndex := 0
	bestDot := dprec.Vec3Dot(h.Points[bestIndex], dir)
	for i, v := range h.Points[1:] {
		if dot := dprec.Vec3Dot(v, dir); dot > bestDot {
			bestDot = dot
			bestIndex = i + 1
		}
	}
	return h.Rotation.Apply(h.Points[bestIndex]), bestIndex
}

// WSPosition returns the world-space position of the hull vertex at the given
// index, taking the hull's rotation into account.
func (h *Hull) WSPosition(index int) dprec.Vec3 {
	return h.Rotation.Apply(h.Points[index])
}
