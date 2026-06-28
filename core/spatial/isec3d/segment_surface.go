package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSegmentSurface reports whether the directed segment crosses the surface
// from front to back.
//
// Like all segment checks in this package, the test is oriented and
// face-culled: the segment is treated as a directed probe from A to B and only
// a crossing where A lies on the side the normal faces and B lies behind it
// counts as an intersection. A back-to-front crossing, or a segment lying in
// the surface, is not considered to intersect it.
func CheckSegmentSurface(segment shape3d.Segment, surface shape3d.Surface) bool {
	distA := surface.SignedDistance(segment.A)
	distB := surface.SignedDistance(segment.B)
	return distA >= 0.0 && distB <= 0.0
}

// ResolveSegmentSurface yields the contact at which the directed segment crosses
// into the surface from the front, if it does.
//
// The contact follows the entry-point convention shared by the segment Resolve
// routines in this package:
//
//   - TargetPoint is the point where the segment crosses the surface.
//   - TargetNormal is the surface normal.
//   - Depth is the fraction of the segment lying beyond the entry point, in the
//     range [0, 1] (1 when the segment enters at A, 0 when it enters at B). It
//     is comparable across shapes, so [shape3d.DeepestContact] selects the
//     earliest entry along the segment.
func ResolveSegmentSurface(segment shape3d.Segment, surface shape3d.Surface, yield shape3d.ContactCallback) {
	distA := surface.SignedDistance(segment.A)
	distB := surface.SignedDistance(segment.B)

	if distA < 0.0 || distB > 0.0 {
		return // not a front-to-back crossing
	}
	denom := distA - distB

	var tContact float64
	if denom == 0 {
		tContact = 0.0 // degenerate, segment lies in the surface
	} else {
		tContact = distA / denom
	}

	contactPoint := dprec.Vec3Lerp(segment.A, segment.B, tContact)

	yield(shape3d.Contact{
		TargetPoint:  contactPoint,
		TargetNormal: surface.Normal,
		Depth:        1.0 - tContact,
	})
}
