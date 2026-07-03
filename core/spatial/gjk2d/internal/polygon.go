package internal

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// Polygon is a convex polygon in local space together with its world-space
// rotation, used by the GJK solver to compute support points.
type Polygon struct {
	// Rotation is the orientation of the polygon in world space.
	Rotation shape2d.Rotation
	// InvRotation is the inverse of Rotation, cached to avoid recomputing it
	// on every support query.
	InvRotation shape2d.Rotation
	// Points holds the local-space vertices of the convex polygon core.
	Points []dprec.Vec2
}

// Support returns the world-space position and the index of the polygon
// vertex that is furthest along dir. The direction is expected to be in
// world space and does not need to be normalized.
func (p *Polygon) Support(dir dprec.Vec2) (dprec.Vec2, int) {
	dir = p.InvRotation.Apply(dir)
	bestIndex := 0
	bestDot := dprec.Vec2Dot(p.Points[bestIndex], dir)
	for i, v := range p.Points[1:] {
		if dot := dprec.Vec2Dot(v, dir); dot > bestDot {
			bestDot = dot
			bestIndex = i + 1
		}
	}
	return p.Rotation.Apply(p.Points[bestIndex]), bestIndex
}
