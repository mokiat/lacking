package query2d

// AABB is an axis-aligned bounding box that can be used for spatial queries.
type AABB struct {
	minX float64
	minY float64
	maxX float64
	maxY float64
}

// NewAABB creates a new [AABB] with the given minimum and maximum coordinates.
func NewAABB(minX, minY, maxX, maxY float64) AABB {
	return AABB{
		minX: minX,
		minY: minY,
		maxX: maxX,
		maxY: maxY,
	}
}

// AABBFromCircle creates an [AABB] that fully contains a circle with the given
// center and radius.
func AABBFromCircle(x, y, r float64) AABB {
	return AABB{
		minX: x - r,
		minY: y - r,
		maxX: x + r,
		maxY: y + r,
	}
}

// AABBFromRectangle creates an [AABB] that fully contains a rectangle with the
// given center and dimensions.
func AABBFromRectangle(x, y, width, height float64) AABB {
	halfWidth := width * 0.5
	halfHeight := height * 0.5
	return AABB{
		minX: x - halfWidth,
		minY: y - halfHeight,
		maxX: x + halfWidth,
		maxY: y + halfHeight,
	}
}

// AABBFromSquare creates an [AABB] that fully contains a square with the given
// center and size.
func AABBFromSquare(x, y, size float64) AABB {
	halfSize := size * 0.5
	return AABB{
		minX: x - halfSize,
		minY: y - halfSize,
		maxX: x + halfSize,
		maxY: y + halfSize,
	}
}
