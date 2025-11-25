package shape2d

// CheckSegmentPolygonIntersection checks if a segment intersects a polygon.
func CheckSegmentPolygonIntersection(segment Segment, polygon Polygon, yield IntersectionYieldFunc) {
	for _, edge := range polygon.Edges {
		CheckSegmentEdgeIntersection(segment, edge, yield)
	}
}
