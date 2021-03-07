package opengl

func NewArea(x, y, width, height int) Area {
	return Area{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

type Area struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (a Area) Empty() bool {
	return a.Width <= 0 && a.Height <= 0
}
