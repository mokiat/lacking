package shape2d

import "github.com/mokiat/gomath/dprec"

// AABB represents an axis-aligned bounding box in 2D space.
type AABB struct {
	// MinX specifies the minimum X coordinate of the box.
	MinX float64
	// MinY specifies the minimum Y coordinate of the box.
	MinY float64
	// MaxX specifies the maximum X coordinate of the box.
	MaxX float64
	// MaxY specifies the maximum Y coordinate of the box.
	MaxY float64
}

// NewAABB creates an [AABB] with the given minimum and maximum coordinates.
func NewAABB(minX, minY, maxX, maxY float64) AABB {
	return AABB{
		MinX: minX,
		MinY: minY,
		MaxX: maxX,
		MaxY: maxY,
	}
}

// AABBFromCircle returns the smallest [AABB] that fully encompasses the
// given circle.
func AABBFromCircle(circle Circle) AABB {
	return AABB{
		MinX: circle.Center.X - circle.Radius,
		MinY: circle.Center.Y - circle.Radius,
		MaxX: circle.Center.X + circle.Radius,
		MaxY: circle.Center.Y + circle.Radius,
	}
}

// AABBFromRectangle returns the smallest [AABB] that fully encompasses the
// given rectangle, taking the rectangle's orientation into account.
func AABBFromRectangle(rect Rectangle) AABB {
	rotation := rect.Rotation
	halfExtentX := dprec.Abs(rotation.BasisX.X)*rect.HalfWidth +
		dprec.Abs(rotation.BasisY.X)*rect.HalfHeight
	halfExtentY := dprec.Abs(rotation.BasisX.Y)*rect.HalfWidth +
		dprec.Abs(rotation.BasisY.Y)*rect.HalfHeight

	return AABB{
		MinX: rect.Center.X - halfExtentX,
		MinY: rect.Center.Y - halfExtentY,
		MaxX: rect.Center.X + halfExtentX,
		MaxY: rect.Center.Y + halfExtentY,
	}
}

// IsEmpty returns whether this AABB has no area, which is the case when its
// minimum coordinate exceeds its maximum coordinate along some axis.
func (aabb AABB) IsEmpty() bool {
	return (aabb.MinX > aabb.MaxX) || (aabb.MinY > aabb.MaxY)
}
