package ui

// Canvas is an interface that represents a mechanism
// through which a Control can render itself to the screen.
type Canvas interface {

	// Push records the current state and creates a new
	// state layer. Changes done in the new layer will
	// not affect the former layer.
	Push()

	// Pop restores the former state layer and configures
	// the drawing state accordingly.
	Pop()

	// Translate moves the drawing position by the specified
	// delta amount.
	Translate(delta Position)

	// Clip sets new clipping bounds. Pixels from draw operations
	// that are outside the clipping bounds will not be drawn.
	//
	// Initially the clipping bounds are equal to the window size.
	Clip(bounds Bounds)

	// SolidColor returns the Color that is used for fill operations.
	SolidColor() Color

	// SetSolidColor sets a new Color to be used for fill operations.
	SetSolidColor(color Color)

	// StrokeColor returns the Color that is used for outline draw
	// operations.
	StrokeColor() Color

	// SetStrokeColor sets a new Color to be used for outline draw
	// operations.
	SetStrokeColor(color Color)

	// StrokeSize returns the size that will be used for outlines.
	StrokeSize() int

	// SetStrokeSize sets a new size to be used for drawing outlines.
	SetStrokeSize(size int)

	// Font returns the Font that is to be used for text draw
	// operations.
	Font() Font

	// SetFont sets a new Font to be used for text draw operations.
	SetFont(font Font)

	// DrawRectangle draws the outlines of a rectangle.
	DrawRectangle(position Position, size Size)

	// FillRectangle draws the solid part of a rectangle.
	FillRectangle(position Position, size Size)

	// DrawRoundRectangle draws the outlines of a rounded rectangle.
	DrawRoundRectangle(position Position, size Size, radius int)

	// FillRoundRectangle draws the solid part of a rounded rectangle.
	FillRoundRectangle(position Position, size Size, radius int)

	// DrawCircle draws the outlines of a circle.
	DrawCircle(position Position, radius int)

	// FillCircle draws the solid part of a circle.
	FillCircle(position Position, radius int)

	// DrawTriangle draws the outlines of a triangle.
	DrawTriangle(a, b, c Position)

	// FillTriangle draws the solid part of a triangle.
	FillTriangle(a, b, c Position)

	// DrawLine draws a line segment.
	DrawLine(start, end Position)

	// DrawImage draws the specified Image.
	DrawImage(image Image, position Position, size Size)

	// DrawText draws a text string.
	DrawText(text string, position Position)
}
