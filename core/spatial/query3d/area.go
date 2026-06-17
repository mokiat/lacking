package query3d

// Area represents the spatial area of an object in the 3D space.
type Area struct {
	x float32
	y float32
	z float32
	r float32
}

// AreaFromSphere creates an area from the given center coordinates and radius.
func AreaFromSphere(x, y, z, r float32) Area {
	return Area{
		x: x,
		y: y,
		z: z,
		r: r,
	}
}

// AreaFromCube creates an area from the given center coordinates and size,
// where the size is the length of the sides of the cubic area.
func AreaFromCube(x, y, z, size float32) Area {
	return Area{
		x: x,
		y: y,
		z: z,
		r: size * 0.5,
	}
}

// AreaFromBox creates an area from the given center coordinates and size,
// where the size is the width, height, and depth of the box area.
func AreaFromBox(x, y, z, width, height, depth float32) Area {
	halfWidth := width * 0.5
	halfHeight := height * 0.5
	halfDepth := depth * 0.5
	return Area{
		x: x,
		y: y,
		z: z,
		r: max(halfWidth, max(halfHeight, halfDepth)),
	}
}
