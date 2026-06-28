package query2d

import "github.com/mokiat/lacking/core/spatial/shape2d"

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

// AABBFromCircle creates an [AABB] that fully contains the given circle.
func AABBFromCircle(circle shape2d.Circle) AABB {
	return AABB{
		minX: circle.Center.X - circle.Radius,
		minY: circle.Center.Y - circle.Radius,
		maxX: circle.Center.X + circle.Radius,
		maxY: circle.Center.Y + circle.Radius,
	}
}

// AABBFromRectangle creates an [AABB] from the given rectangle's center and
// half-extents. The rectangle orientation is ignored, so the result encloses
// the rectangle only when it is axis-aligned.
func AABBFromRectangle(rect shape2d.Rectangle) AABB {
	return AABB{
		minX: rect.Center.X - rect.HalfWidth,
		minY: rect.Center.Y - rect.HalfHeight,
		maxX: rect.Center.X + rect.HalfWidth,
		maxY: rect.Center.Y + rect.HalfHeight,
	}
}
