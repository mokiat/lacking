package isec2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// CheckRectangleEdge reports whether the rectangle intersects the edge.
//
// The edge is treated as a one-sided, front-face-culled boundary exactly as in
// [CheckCircleEdge]: only a rectangle whose center lies strictly in front of the
// edge (on the side the edge's normal faces) and that reaches across the edge's
// line within its span is considered to intersect it. A rectangle centered
// behind the edge, or exactly on its line, is never considered to intersect it.
// A rectangle that merely touches the edge, whether along its span or at an
// endpoint, counts as intersecting.
//
// The rectangle may be arbitrarily oriented; the test is a separating-axis test
// over the edge's normal and the rectangle's two axes.
func CheckRectangleEdge(rectangle shape2d.Rectangle, edge shape2d.Edge) bool {
	delta := dprec.Vec2Diff(edge.B, edge.A)
	if delta.SqrLength() == 0.0 {
		return false // degenerate edge
	}

	axisX := rectangle.Rotation.BasisX
	axisY := rectangle.Rotation.BasisY
	normal := edge.Normal()

	// The edge is one-sided: cull a rectangle whose center is not strictly in
	// front of it, then reject one that does not reach across its line.
	height := dprec.Vec2Dot(normal, dprec.Vec2Diff(rectangle.Center, edge.A))
	if height <= 0.0 {
		return false
	}
	radiusNormal := rectangle.HalfWidth*dprec.Abs(dprec.Vec2Dot(normal, axisX)) +
		rectangle.HalfHeight*dprec.Abs(dprec.Vec2Dot(normal, axisY))
	if radiusNormal-height < 0.0 {
		return false
	}

	// Separating-axis tests against the rectangle's own axes, which bound the
	// edge to its finite span.
	relA := dprec.Vec2Diff(edge.A, rectangle.Center)
	relB := dprec.Vec2Diff(edge.B, rectangle.Center)

	edgeMinX, edgeMaxX := minMax(dprec.Vec2Dot(axisX, relA), dprec.Vec2Dot(axisX, relB))
	if _, _, ok := axisOverlap(rectangle.HalfWidth, edgeMinX, edgeMaxX); !ok {
		return false
	}
	edgeMinY, edgeMaxY := minMax(dprec.Vec2Dot(axisY, relA), dprec.Vec2Dot(axisY, relB))
	if _, _, ok := axisOverlap(rectangle.HalfHeight, edgeMinY, edgeMaxY); !ok {
		return false
	}
	return true
}

// ResolveRectangleEdge yields a [shape2d.Contact] for the overlap of the
// rectangle with the edge, if there is one.
//
// The edge is front-face-culled exactly as in [CheckRectangleEdge], so a
// rectangle centered behind it produces no contact. The contact is reported with
// the rectangle as the source and the edge as the target: TargetNormal is the
// outward unit direction along which the rectangle must be moved by Depth to
// resolve the overlap, and TargetPoint is the corresponding point on the edge.
//
// The contact resolves along the axis of least penetration, mirroring the span
// and endpoint cases of [ResolveCircleEdge]: TargetNormal is the edge's outward
// normal when the rectangle overlaps the edge's span, or one of the rectangle's
// own axes when the rectangle meets the edge near an endpoint.
//
// TargetPoint is the spot on the edge that the penetrating feature (a
// rectangle corner, an edge endpoint, or a face span) touches once the
// rectangle is moved out by Depth: the center of the region where the edge
// meets the rectangle at its just-touching position. It is consistent with
// TargetNormal and Depth, so [shape2d.Contact.EvalSourcePoint] evaluates to
// the matching point on the rectangle.
func ResolveRectangleEdge(rectangle shape2d.Rectangle, edge shape2d.Edge, yield shape2d.ContactCallback) {
	delta := dprec.Vec2Diff(edge.B, edge.A)
	sqrLength := delta.SqrLength()
	if sqrLength == 0.0 {
		return // degenerate edge
	}

	axisX := rectangle.Rotation.BasisX
	axisY := rectangle.Rotation.BasisY
	normal := edge.Normal()
	tangent := dprec.UnitVec2(delta)
	length := dprec.Sqrt(sqrLength)

	// Edge normal axis. The edge is one-sided, so the rectangle is only ever
	// pushed toward the front (along the edge normal); a center behind the edge
	// or a rectangle that does not reach its line yields no contact.
	height := dprec.Vec2Dot(normal, dprec.Vec2Diff(rectangle.Center, edge.A))
	if height <= 0.0 {
		return
	}
	radiusNormal := rectangle.HalfWidth*dprec.Abs(dprec.Vec2Dot(normal, axisX)) +
		rectangle.HalfHeight*dprec.Abs(dprec.Vec2Dot(normal, axisY))
	penetrationNormal := radiusNormal - height
	if penetrationNormal < 0.0 {
		return
	}

	// Rectangle axes. These separating-axis tests bound the edge to its finite
	// span; a non-overlap on either means the rectangle slips past an endpoint.
	relA := dprec.Vec2Diff(edge.A, rectangle.Center)
	relB := dprec.Vec2Diff(edge.B, rectangle.Center)

	projAX := dprec.Vec2Dot(axisX, relA)
	projBX := dprec.Vec2Dot(axisX, relB)
	edgeMinX, edgeMaxX := minMax(projAX, projBX)
	penetrationX, signX, ok := axisOverlap(rectangle.HalfWidth, edgeMinX, edgeMaxX)
	if !ok {
		return
	}
	projAY := dprec.Vec2Dot(axisY, relA)
	projBY := dprec.Vec2Dot(axisY, relB)
	edgeMinY, edgeMaxY := minMax(projAY, projBY)
	penetrationY, signY, ok := axisOverlap(rectangle.HalfHeight, edgeMinY, edgeMaxY)
	if !ok {
		return
	}

	// Pick the axis of least penetration, preferring the edge normal on ties so
	// that a rectangle squarely on the span resolves along the boundary.
	contactNormal := normal
	depth := penetrationNormal
	if penetrationX < depth {
		depth = penetrationX
		contactNormal = dprec.Vec2Prod(axisX, signX)
	}
	if penetrationY < depth {
		depth = penetrationY
		contactNormal = dprec.Vec2Prod(axisY, signY)
	}

	// The contact point is the center of the touch region: moving the rectangle
	// by depth along the normal brings the two shapes to a just-touching
	// position, and the edge is clipped against the rectangle at that position
	// (inflated slightly for robustness). Because the point lies on the
	// translated rectangle, stepping back by Depth along the normal, as
	// [shape2d.Contact.EvalSourcePoint] does, lands on the original rectangle.
	touchOffsetX := depth * dprec.Vec2Dot(contactNormal, axisX)
	touchOffsetY := depth * dprec.Vec2Dot(contactNormal, axisY)
	inflation := 1e-7 * (1.0 + rectangle.HalfWidth + rectangle.HalfHeight)

	overlapStart, overlapEnd := 0.0, 1.0
	hasOverlap := true
	if low, high, ok := slabRange(projAX-touchOffsetX, projBX-projAX, rectangle.HalfWidth+inflation); ok {
		overlapStart, overlapEnd = max(overlapStart, low), min(overlapEnd, high)
	} else {
		hasOverlap = false
	}
	if low, high, ok := slabRange(projAY-touchOffsetY, projBY-projAY, rectangle.HalfHeight+inflation); ok {
		overlapStart, overlapEnd = max(overlapStart, low), min(overlapEnd, high)
	} else {
		hasOverlap = false
	}
	hasOverlap = hasOverlap && (overlapStart <= overlapEnd)

	var contactPoint dprec.Vec2
	if hasOverlap {
		contactPoint = dprec.Vec2Lerp(edge.A, edge.B, (overlapStart+overlapEnd)/2.0)
	} else {
		// The clip came up empty (a grazing touch lost to floating point): fall
		// back to the deepest point of the rectangle, projected onto the edge.
		support := rectangle.Center
		support = dprec.Vec2Diff(support, dprec.Vec2Prod(axisX, rectangle.HalfWidth*supportSign(dprec.Vec2Dot(contactNormal, axisX))))
		support = dprec.Vec2Diff(support, dprec.Vec2Prod(axisY, rectangle.HalfHeight*supportSign(dprec.Vec2Dot(contactNormal, axisY))))
		projected := dprec.Vec2Sum(support, dprec.Vec2Prod(contactNormal, depth))
		span := dprec.Vec2Dot(dprec.Vec2Diff(projected, edge.A), tangent)
		span = min(max(span, 0.0), length)
		contactPoint = dprec.Vec2Sum(edge.A, dprec.Vec2Prod(tangent, span))
	}

	yield(shape2d.Contact{
		TargetPoint:  contactPoint,
		TargetNormal: contactNormal,
		Depth:        depth,
	})
}

// supportSign returns the sign of the given axis projection, treating
// near-perpendicular projections as zero so that flat support features (faces)
// resolve to their midpoint rather than an arbitrary corner.
func supportSign(projection float64) float64 {
	const epsilon = 1e-9
	switch {
	case projection > epsilon:
		return 1.0
	case projection < -epsilon:
		return -1.0
	default:
		return 0.0
	}
}

// minMax returns its two arguments ordered as the smaller and the larger.
func minMax(a, b float64) (float64, float64) {
	if a < b {
		return a, b
	}
	return b, a
}

// axisOverlap measures the overlap, on a single separating axis, of a rectangle
// centered at the origin of that axis with the given projection radius against
// the edge interval [edgeMin, edgeMax].
//
// It returns the penetration depth and the sign (+1 or -1) of the axis direction
// along which the rectangle must be moved to resolve that overlap. ok is false
// when the projections are disjoint.
func axisOverlap(radius, edgeMin, edgeMax float64) (depth, sign float64, ok bool) {
	rectMin := -radius
	rectMax := radius
	penetrationPositive := edgeMax - rectMin // move the rectangle along +axis
	penetrationNegative := rectMax - edgeMin // move the rectangle along -axis
	if penetrationPositive < 0.0 || penetrationNegative < 0.0 {
		return 0.0, 0.0, false
	}
	if penetrationPositive < penetrationNegative {
		return penetrationPositive, 1.0, true
	}
	return penetrationNegative, -1.0, true
}
