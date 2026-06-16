package query2d

// AABB is an axis-aligned bounding box that can be used for spatial queries.
type AABB struct {
	minX float32
	minY float32
	maxX float32
	maxY float32
}

// NewAABB creates a new AABB with the given minimum and maximum coordinates.
func NewAABB(minX, minY, maxX, maxY float32) AABB {
	return AABB{
		minX: minX,
		minY: minY,
		maxX: maxX,
		maxY: maxY,
	}
}

// AABBFromCircle creates an AABB that fully contains a circle with the given
// center and radius.
func AABBFromCircle(x, y, r float32) AABB {
	return AABB{
		minX: x - r,
		minY: y - r,
		maxX: x + r,
		maxY: y + r,
	}
}
