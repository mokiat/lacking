package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSegmentSurface reports whether the segment crosses the surface.
//
// The test is symmetric: it does not depend on the order of the segment's
// endpoints nor on which side of the surface they lie. A segment that merely
// touches the surface with one endpoint is considered to intersect it.
func CheckSegmentSurface(segment shape3d.Segment, surface shape3d.Surface) bool {
	distA := surface.SignedDistance(segment.A)
	distB := surface.SignedDistance(segment.B)
	return distA*distB <= 0.0
}

// ResolveSegmentSurface yields a Contact for the crossing of the segment with
// the surface, if there is one.
//
// The contact is expressed relative to the surface as the target shape:
// TargetPoint is the point where the segment crosses the surface, TargetNormal
// is the surface normal, and Depth is how far the deepest endpoint lies behind
// the surface, measured along the normal.
func ResolveSegmentSurface(segment shape3d.Segment, surface shape3d.Surface, yield shape3d.ContactCallback) {
	distA := surface.SignedDistance(segment.A)
	distB := surface.SignedDistance(segment.B)
	if distA*distB > 0.0 {
		return // both endpoints on the same side; no crossing
	}

	denom := distA - distB
	var lerpFactor float64
	if denom != 0 {
		lerpFactor = distA / denom
	}
	contactPoint := dprec.Vec3Lerp(segment.A, segment.B, lerpFactor)

	yield(shape3d.Contact{
		TargetPoint:  contactPoint,
		TargetNormal: surface.Normal,
		Depth:        -min(distA, distB),
	})
}
