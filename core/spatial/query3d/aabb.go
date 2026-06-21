package query3d

// AABB is an axis-aligned bounding box that can be used for spatial queries.
type AABB struct {
	minX float64
	minY float64
	minZ float64
	maxX float64
	maxY float64
	maxZ float64
}

// NewAABB creates a new AABB with the given minimum and maximum coordinates.
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

// AABBFromSphere creates an AABB that fully contains a sphere with the given
// center and radius.
func AABBFromSphere(x, y, z, r float64) AABB {
	return AABB{
		minX: x - r,
		minY: y - r,
		minZ: z - r,
		maxX: x + r,
		maxY: y + r,
		maxZ: z + r,
	}
}

// AABBFromBox creates an AABB that fully contains a box with the given
// center and dimensions.
func AABBFromBox(x, y, z, width, height, depth float64) AABB {
	halfWidth := width * 0.5
	halfHeight := height * 0.5
	halfDepth := depth * 0.5
	return AABB{
		minX: x - halfWidth,
		minY: y - halfHeight,
		minZ: z - halfDepth,
		maxX: x + halfWidth,
		maxY: y + halfHeight,
		maxZ: z + halfDepth,
	}
}

// AABBFromCube creates an AABB that fully contains a cube with the given
// center and size.
func AABBFromCube(x, y, z, size float64) AABB {
	halfSize := size * 0.5
	return AABB{
		minX: x - halfSize,
		minY: y - halfSize,
		minZ: z - halfSize,
		maxX: x + halfSize,
		maxY: y + halfSize,
		maxZ: z + halfSize,
	}
}
