package shape2d

// TODO: Write test for this!

// CheckSegmentPolygonIntersection checks if a segment intersects a polygon.
func CheckSegmentPolygonIntersection(segment Segment, polygon Polygon) (Intersection, bool) {
	var collection FarthestIntersection
	for _, polySegment := range polygon.Segments {
		if intersection, ok := CheckSegmentSegmentIntersection(segment, polySegment); ok {
			collection.AddIntersection(intersection)
		}
	}
	return collection.Intersection()
}
