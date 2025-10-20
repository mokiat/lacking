package shape2d

import "github.com/mokiat/gomath/dprec"

// CheckCircleEdgeIntersection checks for intersection between a circle and an
// edge. The orientation of the edge matters.
func CheckCircleEdgeIntersection(circle Circle, edge Edge) (Intersection, bool) {
	vecAB := dprec.Vec2Diff(edge.B, edge.A)
	vecAC := dprec.Vec2Diff(circle.Position, edge.A)
	vecBC := dprec.Vec2Diff(circle.Position, edge.B)

	normal := edge.Normal()
	height := dprec.Vec2Dot(normal, vecAC)
	if height > circle.Radius || height <= 0.0 {
		return Intersection{}, false
	}

	u := dprec.Vec2Dot(vecAC, vecAB) / vecAB.SqrLength()
	switch {
	case u < 0.0:
		sqrDistance := vecAC.SqrLength()
		if sqrDistance > circle.Radius*circle.Radius {
			return Intersection{}, false
		}
		distance := dprec.Sqrt(sqrDistance)
		normal := dprec.Vec2Quot(vecAC, distance)
		return Intersection{
			TargetContact: edge.A,
			TargetNormal:  normal,
			Depth:         circle.Radius - distance,
		}, true

	case u > 1.0:
		sqrDistance := vecBC.SqrLength()
		if sqrDistance > circle.Radius*circle.Radius {
			return Intersection{}, false
		}
		distance := dprec.Sqrt(sqrDistance)
		normal := dprec.Vec2Quot(vecBC, distance)
		return Intersection{
			TargetContact: edge.B,
			TargetNormal:  normal,
			Depth:         circle.Radius - distance,
		}, true

	default:
		projection := dprec.Vec2Lerp(edge.A, edge.B, u)
		return Intersection{
			TargetContact: projection,
			TargetNormal:  normal,
			Depth:         circle.Radius - height,
		}, true
	}
}
