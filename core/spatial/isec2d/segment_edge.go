package isec2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// CheckSegmentEdge reports whether the directed segment crosses the edge.
//
// Like all segment checks in this package, the test is oriented and
// face-culled. The segment is treated as a directed probe from A to B and the
// edge as a one-sided boundary whose normal (see [shape2d.Edge.Normal]) marks
// its front side. Only a segment that crosses the edge's span while moving from
// the front side toward the back, within its own extent, counts as an
// intersection. A segment travelling the other way (from behind the edge toward
// its front), one running parallel to the edge, or one whose crossing of the
// edge's supporting line falls outside the A-to-B span or outside the segment's
// extent, is not considered to intersect it.
func CheckSegmentEdge(segment shape2d.Segment, edge shape2d.Edge) bool {
	vecAC := dprec.Vec2Diff(segment.A, edge.A)
	vecAB := dprec.Vec2Diff(edge.B, edge.A)
	vecDC := dprec.Vec2Diff(segment.A, segment.B)

	det := dprec.Vec2Cross(vecDC, vecAB)
	if det <= 0.0 {
		return false
	}

	detU := dprec.Vec2Cross(vecDC, vecAC)
	if detU < 0.0 || detU > det {
		return false
	}

	detV := dprec.Vec2Cross(vecAC, vecAB)
	if detV < 0.0 || detV > det {
		return false
	}

	return true
}

// ResolveSegmentEdge yields the contact at which the directed segment crosses
// the edge, if it crosses it at all.
//
// The crossing is oriented and face-culled exactly as in [CheckSegmentEdge], so
// only a front-to-back crossing within both the edge's span and the segment's
// extent produces a contact. The contact follows the entry-point convention
// shared by the segment Resolve routines in this package:
//
//   - TargetPoint is the point where the segment crosses the edge.
//   - TargetNormal is the edge's outward normal (see [shape2d.Edge.Normal]),
//     which faces back toward the side the segment came from.
//   - Depth is the fraction of the segment lying beyond the crossing, in the
//     range [0, 1] (1 when the segment crosses at A, 0 when it crosses at B). It
//     is comparable across shapes, so [shape2d.DeepestContact] selects the
//     earliest crossing along the segment.
func ResolveSegmentEdge(segment shape2d.Segment, edge shape2d.Edge, yield shape2d.ContactCallback) {
	vecAC := dprec.Vec2Diff(segment.A, edge.A)
	vecAB := dprec.Vec2Diff(edge.B, edge.A)
	vecDC := dprec.Vec2Diff(segment.A, segment.B)

	det := dprec.Vec2Cross(vecDC, vecAB)
	if det <= 0.0 {
		return
	}

	detU := dprec.Vec2Cross(vecDC, vecAC)
	if detU < 0.0 || detU > det {
		return
	}

	detV := dprec.Vec2Cross(vecAC, vecAB)
	if detV < 0.0 || detV > det {
		return
	}

	tContact := detV / det
	contactPoint := dprec.Vec2Lerp(segment.A, segment.B, tContact)
	normal := edge.Normal()

	yield(shape2d.Contact{
		TargetPoint:  contactPoint,
		TargetNormal: normal,
		Depth:        1.0 - tContact,
	})
}
