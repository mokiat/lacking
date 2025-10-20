package shape2d

// CheckCirclePolygonIntersection checks for intersection between a circle and
// a polygon.
func CheckCirclePolygonIntersection(circle Circle, polygon Polygon) (Intersection, bool) {
	var collection LargestIntersection
	for _, edge := range polygon.Edges {
		if !IsCircleCircleIntersection(circle, edge.BoundingCircle()) {
			continue
		}
		if intersection, ok := CheckCircleEdgeIntersection(circle, edge); ok {
			collection.AddIntersection(intersection)
		}
	}
	return collection.Intersection()
}
