package shape2d

import "github.com/mokiat/gomath/dprec"

// IsCircleEdgeIntersection checks if a circle intersects with an edge.
// The orientation of the edge matters.
//
// Only a bool result is returned and no collision points or separation
// normals are evaluated.
func IsCircleEdgeIntersection(circle Circle, edge Edge) bool {
	vecAB := dprec.Vec2Diff(edge.B, edge.A)
	vecAC := dprec.Vec2Diff(circle.Position, edge.A)
	vecBC := dprec.Vec2Diff(circle.Position, edge.B)

	normal := edge.Normal()
	height := dprec.Vec2Dot(normal, vecAC)
	if height > circle.Radius || height <= 0.0 {
		return false
	}

	u := dprec.Vec2Dot(vecAC, vecAB) / vecAB.SqrLength()
	switch {
	case u < 0.0:
		sqrDistance := vecAC.SqrLength()
		return sqrDistance <= circle.Radius*circle.Radius && sqrDistance >= 0.0001*0.0001
	case u > 1.0:
		sqrDistance := vecBC.SqrLength()
		return sqrDistance <= circle.Radius*circle.Radius && sqrDistance >= 0.0001*0.0001
	default:
		return true
	}
}

// CheckCircleEdgeIntersection checks for intersection between a circle and an
// edge. The orientation of the edge matters.
func CheckCircleEdgeIntersection(circle Circle, edge Edge, yield IntersectionYieldFunc) {
	vecAB := dprec.Vec2Diff(edge.B, edge.A)
	vecAC := dprec.Vec2Diff(circle.Position, edge.A)
	vecBC := dprec.Vec2Diff(circle.Position, edge.B)

	normal := edge.Normal()
	height := dprec.Vec2Dot(normal, vecAC)
	if height > circle.Radius || height <= 0.0 {
		return
	}

	u := dprec.Vec2Dot(vecAC, vecAB) / vecAB.SqrLength()
	switch {
	case u < 0.0:
		sqrDistance := vecAC.SqrLength()
		if sqrDistance > circle.Radius*circle.Radius {
			return
		}
		if sqrDistance < 0.0001*0.0001 {
			return
		}
		distance := dprec.Sqrt(sqrDistance)
		normal := dprec.Vec2Quot(vecAC, distance)
		yield(Intersection{
			TargetContact: edge.A,
			TargetNormal:  normal,
			Depth:         circle.Radius - distance,
		})

	case u > 1.0:
		sqrDistance := vecBC.SqrLength()
		if sqrDistance > circle.Radius*circle.Radius {
			return
		}
		if sqrDistance < 0.0001*0.0001 {
			return
		}
		distance := dprec.Sqrt(sqrDistance)
		normal := dprec.Vec2Quot(vecBC, distance)
		yield(Intersection{
			TargetContact: edge.B,
			TargetNormal:  normal,
			Depth:         circle.Radius - distance,
		})

	default:
		projection := dprec.Vec2Lerp(edge.A, edge.B, u)
		yield(Intersection{
			TargetContact: projection,
			TargetNormal:  normal,
			Depth:         circle.Radius - height,
		})
	}
}
