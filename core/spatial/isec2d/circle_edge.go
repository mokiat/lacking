package isec2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// CheckCircleEdge reports whether the circle intersects the edge.
//
// The edge is treated as a one-sided, front-face-culled boundary: only a circle
// whose center lies strictly in front of the edge (on the side the edge's normal
// faces) and within the circle radius of the edge's line can intersect it. A
// circle centered behind the edge, or exactly on its line, is never considered
// to intersect it. A circle whose center is in front and that merely touches the
// edge, whether along its span or at an endpoint, counts as intersecting.
func CheckCircleEdge(circle shape2d.Circle, edge shape2d.Edge) bool {
	vecAB := dprec.Vec2Diff(edge.B, edge.A)
	vecAC := dprec.Vec2Diff(circle.Center, edge.A)
	vecBC := dprec.Vec2Diff(circle.Center, edge.B)

	normal := edge.Normal()
	height := dprec.Vec2Dot(normal, vecAC)
	if height > circle.Radius || height <= 0.0 {
		return false
	}

	sqrEdgeLength := vecAB.SqrLength()
	if sqrEdgeLength == 0.0 {
		return false
	}

	u := dprec.Vec2Dot(vecAC, vecAB) / sqrEdgeLength
	switch {
	case u < 0.0:
		sqrDistance := vecAC.SqrLength()
		return sqrDistance <= circle.Radius*circle.Radius
	case u > 1.0:
		sqrDistance := vecBC.SqrLength()
		return sqrDistance <= circle.Radius*circle.Radius
	default:
		return true
	}
}

// ResolveCircleEdge yields a [shape2d.Contact] for the overlap of the circle
// with the edge, if there is one.
//
// The edge is front-face-culled exactly as in [CheckCircleEdge], so a circle
// centered behind it produces no contact. The contact is reported with the
// circle as the source and the edge as the target: TargetPoint is the point on
// the edge closest to the circle center (along its span or at an endpoint),
// TargetNormal is the outward unit direction from that point toward the circle
// center, and Depth is how far the circle penetrates, that is the circle radius
// minus the distance to that closest point. Moving the circle by Depth along
// TargetNormal resolves the overlap.
func ResolveCircleEdge(circle shape2d.Circle, edge shape2d.Edge, yield shape2d.ContactCallback) {
	vecAB := dprec.Vec2Diff(edge.B, edge.A)
	vecAC := dprec.Vec2Diff(circle.Center, edge.A)
	vecBC := dprec.Vec2Diff(circle.Center, edge.B)

	normal := edge.Normal()
	height := dprec.Vec2Dot(normal, vecAC)
	if height > circle.Radius || height <= 0.0 {
		return
	}

	sqrEdgeLength := vecAB.SqrLength()
	if sqrEdgeLength == 0.0 {
		return
	}

	u := dprec.Vec2Dot(vecAC, vecAB) / sqrEdgeLength
	switch {
	case u < 0.0:
		sqrDistance := vecAC.SqrLength()
		if sqrDistance > circle.Radius*circle.Radius {
			return
		}
		distance := dprec.Sqrt(sqrDistance)
		normal := dprec.Vec2Quot(vecAC, distance)
		yield(shape2d.Contact{
			TargetPoint:  edge.A,
			TargetNormal: normal,
			Depth:        circle.Radius - distance,
		})

	case u > 1.0:
		sqrDistance := vecBC.SqrLength()
		if sqrDistance > circle.Radius*circle.Radius {
			return
		}
		distance := dprec.Sqrt(sqrDistance)
		normal := dprec.Vec2Quot(vecBC, distance)
		yield(shape2d.Contact{
			TargetPoint:  edge.B,
			TargetNormal: normal,
			Depth:        circle.Radius - distance,
		})

	default:
		projection := dprec.Vec2Lerp(edge.A, edge.B, u)
		yield(shape2d.Contact{
			TargetPoint:  projection,
			TargetNormal: normal,
			Depth:        circle.Radius - height,
		})
	}
}
