package shape2d

// CheckCirclePolygonIntersection checks for intersection between a circle and
// a polygon.
func CheckCirclePolygonIntersection(circle Circle, polygon Polygon, yield IntersectionYieldFunc) {
	var collection LargestIntersection
	for _, edge := range polygon.Edges {
		if !IsCircleCircleIntersection(circle, edge.BoundingCircle()) {
			continue
		}
		CheckCircleEdgeIntersection(circle, edge, collection.AddIntersection)
	}
	if intersection, ok := collection.Intersection(); ok {
		yield(intersection)
	}
}
