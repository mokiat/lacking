package shape2d

import "github.com/mokiat/gomath/dprec"

// CheckRectanglePolygonIntersection checks if a Rectangle shape intersects
// with a Polygon shape.
func CheckRectanglePolygonIntersection(rectangle Rectangle, polygon Polygon, yield IntersectionYieldFunc) {
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

	for _, edge := range polygon.Edges {
		CheckSegmentEdgeIntersection(NewSegment(p1, p2), edge, yield)
		CheckSegmentEdgeIntersection(NewSegment(p2, p3), edge, yield)
		CheckSegmentEdgeIntersection(NewSegment(p3, p4), edge, yield)
		CheckSegmentEdgeIntersection(NewSegment(p4, p1), edge, yield)

		// since segment intersections are unidirectional, check the opposite direction as well

		CheckSegmentEdgeIntersection(NewSegment(p2, p1), edge, yield)
		CheckSegmentEdgeIntersection(NewSegment(p3, p2), edge, yield)
		CheckSegmentEdgeIntersection(NewSegment(p4, p3), edge, yield)
		CheckSegmentEdgeIntersection(NewSegment(p1, p4), edge, yield)
	}
}
