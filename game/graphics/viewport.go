package graphics

// NewViewport creates a new Viewport with the specified
// parameters.
func NewViewport(x, y, width, height int) Viewport {
	return Viewport{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// Viewport represents an area on the screen to
// which rendering will occur.
type Viewport struct {
	X      int
	Y      int
	Width  int
	Height int
}
