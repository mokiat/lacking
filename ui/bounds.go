package ui

// Bounds represents an Element's relative
// placing.
type Bounds struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (b Bounds) Translate(dX, dY int) Bounds {
	return Bounds{
		X:      b.X + dX,
		Y:      b.Y + dY,
		Width:  b.Width,
		Height: b.Height,
	}
}

func (b Bounds) Intersect(other Bounds) Bounds {
	x := maxInt(b.X, other.X)
	y := maxInt(b.Y, other.Y)
	width := minInt(b.X+b.Width, other.X+other.Width) - x
	height := minInt(b.Y+b.Height, other.Y+other.Height) - y
	return Bounds{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

func (b Bounds) Empty() bool {
	return b.Width <= 0 || b.Height <= 0
}

// Spacing represents an amount of spacing
// around or inside an Element.
type Spacing struct {
	Left   int
	Right  int
	Top    int
	Bottom int
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
