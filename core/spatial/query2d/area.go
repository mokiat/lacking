package query2d

// Area represents the spatial area of an object in the 2D space.
type Area struct {
	x float32
	y float32
	r float32
}

// CircularArea creates an area from the given center coordinates and radius.
func CircularArea(x, y, r float32) Area {
	return Area{
		x: x,
		y: y,
		r: r,
	}
}

// SquareArea creates an area from the given center coordinates and size,
// where the size is the length of the sides of the square area.
func SquareArea(x, y, size float32) Area {
	return Area{
		x: x,
		y: y,
		r: size * 0.5,
	}
}
