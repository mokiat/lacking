package query3d

import "github.com/mokiat/lacking/core/spatial/shape3d"

// AABB is an axis-aligned bounding box that can be used for spatial queries.
type AABB struct {
	minX float64
	minY float64
	minZ float64
	maxX float64
	maxY float64
	maxZ float64
}

// NewAABB creates a new [AABB] with the given minimum and maximum coordinates.
func NewAABB(minX, minY, minZ, maxX, maxY, maxZ float64) AABB {
	return AABB{
		minX: minX,
		minY: minY,
		minZ: minZ,
		maxX: maxX,
		maxY: maxY,
		maxZ: maxZ,
	}
}

// AABBFromSphere creates an [AABB] that fully contains the given sphere.
func AABBFromSphere(sphere shape3d.Sphere) AABB {
	return AABB{
		minX: sphere.Center.X - sphere.Radius,
		minY: sphere.Center.Y - sphere.Radius,
		minZ: sphere.Center.Z - sphere.Radius,
		maxX: sphere.Center.X + sphere.Radius,
		maxY: sphere.Center.Y + sphere.Radius,
		maxZ: sphere.Center.Z + sphere.Radius,
	}
}

// AABBFromBox creates an [AABB] from the given box's center and half-extents.
// The box orientation is ignored, so the result encloses the box only when it
// is axis-aligned.
func AABBFromBox(box shape3d.Box) AABB {
	return AABB{
		minX: box.Center.X - box.HalfWidth,
		minY: box.Center.Y - box.HalfHeight,
		minZ: box.Center.Z - box.HalfLength,
		maxX: box.Center.X + box.HalfWidth,
		maxY: box.Center.Y + box.HalfHeight,
		maxZ: box.Center.Z + box.HalfLength,
	}
}
