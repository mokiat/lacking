package query2d

// Area represents the spatial area of an object in the 2D space.
type Area struct {
	x float64
	y float64
	r float64
}

// AreaFromCircle creates an area from the given center coordinates and radius.
func AreaFromCircle(x, y, r float64) Area {
	return Area{
		x: x,
		y: y,
		r: r,
	}
}

// AreaFromRectangle creates an area from the given center coordinates and size,
// where the size is the width and height of the rectangular area.
func AreaFromRectangle(x, y, width, height float64) Area {
	halfWidth := width * 0.5
	halfHeight := height * 0.5
	return Area{
		x: x,
		y: y,
		r: max(halfWidth, halfHeight),
	}
}
