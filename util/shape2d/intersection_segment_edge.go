package shape2d

import "github.com/mokiat/gomath/dprec"

// CheckSegmentEdgeIntersection checks if a segment intersects a polygon edge.
func CheckSegmentEdgeIntersection(segment Segment, edge Edge, yield IntersectionYieldFunc) {
	vecAC := dprec.Vec2Diff(segment.A, edge.A)    // offset
	vecAB := dprec.Vec2Diff(edge.B, edge.A)       // edge
	vecDC := dprec.Vec2Diff(segment.A, segment.B) // inverse segment

	det := dprec.Vec2Cross(vecDC, vecAB)
	if det < 0.00001 {
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

	contactPoint := dprec.Vec2Lerp(segment.A, segment.B, detV/det)
	normal := edge.Normal()
	depth := dprec.Vec2Dot(
		dprec.Vec2Diff(contactPoint, segment.B),
		normal,
	)

	yield(Intersection{
		TargetContact: contactPoint,
		TargetNormal:  normal,
		Depth:         depth,
	})
}
