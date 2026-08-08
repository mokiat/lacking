package shape3d

import "github.com/mokiat/gomath/dprec"

// AABB represents an axis-aligned bounding box in 3D space.
type AABB struct {
	// MinX specifies the minimum X coordinate of the box.
	MinX float64
	// MinY specifies the minimum Y coordinate of the box.
	MinY float64
	// MinZ specifies the minimum Z coordinate of the box.
	MinZ float64
	// MaxX specifies the maximum X coordinate of the box.
	MaxX float64
	// MaxY specifies the maximum Y coordinate of the box.
	MaxY float64
	// MaxZ specifies the maximum Z coordinate of the box.
	MaxZ float64
}

// NewAABB creates an [AABB] with the given minimum and maximum coordinates.
func NewAABB(minX, minY, minZ, maxX, maxY, maxZ float64) AABB {
	return AABB{
		MinX: minX,
		MinY: minY,
		MinZ: minZ,
		MaxX: maxX,
		MaxY: maxY,
		MaxZ: maxZ,
	}
}

// AABBFromSphere returns the smallest [AABB] that fully encompasses the
// given sphere.
func AABBFromSphere(sphere Sphere) AABB {
	return AABB{
		MinX: sphere.Center.X - sphere.Radius,
		MinY: sphere.Center.Y - sphere.Radius,
		MinZ: sphere.Center.Z - sphere.Radius,
		MaxX: sphere.Center.X + sphere.Radius,
		MaxY: sphere.Center.Y + sphere.Radius,
		MaxZ: sphere.Center.Z + sphere.Radius,
	}
}

// AABBFromBox returns the smallest [AABB] that fully encompasses the given
// box, taking the box's orientation into account.
func AABBFromBox(box Box) AABB {
	rotation := box.Rotation
	halfExtentX := dprec.Abs(rotation.BasisX.X)*box.HalfWidth +
		dprec.Abs(rotation.BasisY.X)*box.HalfHeight +
		dprec.Abs(rotation.BasisZ.X)*box.HalfLength
	halfExtentY := dprec.Abs(rotation.BasisX.Y)*box.HalfWidth +
		dprec.Abs(rotation.BasisY.Y)*box.HalfHeight +
		dprec.Abs(rotation.BasisZ.Y)*box.HalfLength
	halfExtentZ := dprec.Abs(rotation.BasisX.Z)*box.HalfWidth +
		dprec.Abs(rotation.BasisY.Z)*box.HalfHeight +
		dprec.Abs(rotation.BasisZ.Z)*box.HalfLength

	return AABB{
		MinX: box.Center.X - halfExtentX,
		MinY: box.Center.Y - halfExtentY,
		MinZ: box.Center.Z - halfExtentZ,
		MaxX: box.Center.X + halfExtentX,
		MaxY: box.Center.Y + halfExtentY,
		MaxZ: box.Center.Z + halfExtentZ,
	}
}

// IsEmpty returns whether this AABB has no volume, which is the case when
// its minimum coordinate exceeds its maximum coordinate along some axis.
func (aabb AABB) IsEmpty() bool {
	return (aabb.MinX > aabb.MaxX) || (aabb.MinY > aabb.MaxY) || (aabb.MinZ > aabb.MaxZ)
}
