package shape2d

// IsCirclePolygonIntersection checks if a circle intersects with a polygon.
//
// Only a bool result is returned and no collision points or separation
// normals are evaluated.
func IsCirclePolygonIntersection(circle Circle, polygon Polygon) bool {
	for _, edge := range polygon.Edges {
		if !IsCircleCircleIntersection(circle, edge.BoundingCircle()) {
			continue
		}
		if IsCircleEdgeIntersection(circle, edge) {
			return true
		}
	}
	return false
}

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
