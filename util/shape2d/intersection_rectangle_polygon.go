package shape2d

import "github.com/mokiat/gomath/dprec"

// CheckRectanglePolygonIntersection checks if a Rectangle shape intersects
// with a Polygon shape.
func CheckRectanglePolygonIntersection(rectangle Rectangle, polygon Polygon) (Intersection, bool) {
	rectanglePosition := rectangle.Position
	rectangleRotation := rectangle.Rotation

	orientationX := dprec.AngleVec2Rotation(rectangleRotation, dprec.BasisXVec2())
	orientationY := dprec.AngleVec2Rotation(rectangleRotation, dprec.BasisYVec2())

	maxX := dprec.Vec2Prod(orientationX, rectangle.HalfWidth)
	minX := dprec.InverseVec2(maxX)
	maxY := dprec.Vec2Prod(orientationY, rectangle.HalfHeight)
	minY := dprec.InverseVec2(maxY)

	p1 := dprec.Vec2Sum(dprec.Vec2Sum(rectanglePosition, minX), maxY)
	p2 := dprec.Vec2Sum(dprec.Vec2Sum(rectanglePosition, minX), minY)
	p3 := dprec.Vec2Sum(dprec.Vec2Sum(rectanglePosition, maxX), minY)
	p4 := dprec.Vec2Sum(dprec.Vec2Sum(rectanglePosition, maxX), maxY)

	var bestIntersection SmallestIntersection
	for _, edge := range polygon.Edges {
		if intersection, ok := CheckSegmentEdgeIntersection(NewSegment(p1, p2), edge); ok {
			bestIntersection.AddIntersection(intersection)
		}
		if intersection, ok := CheckSegmentEdgeIntersection(NewSegment(p2, p3), edge); ok {
			bestIntersection.AddIntersection(intersection)
		}
		if intersection, ok := CheckSegmentEdgeIntersection(NewSegment(p3, p4), edge); ok {
			bestIntersection.AddIntersection(intersection)
		}
		if intersection, ok := CheckSegmentEdgeIntersection(NewSegment(p4, p1), edge); ok {
			bestIntersection.AddIntersection(intersection)
		}
	}
	return bestIntersection.Intersection()
}
