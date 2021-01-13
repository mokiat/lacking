package ui

type Canvas interface {
	Push()
	Pop()

	Translate(delta Position)
	Clip(bounds Bounds)

	// SetTypeface(font Typeface)
	// Typeface() Typeface

	// SetStrokeActive(enabled bool)
	// IsStrokeActive() bool

	// SetStrokeColor(color Color)
	// StrokeColor() Color

	// SetStrokeSize(size int)
	// StrokeSize() int

	// SetFillActive(active bool)
	// IsFillActive() bool

	// SetFillColor(color Color)
	// FillColor() Color

	// DrawRectangle(position Position, size Size)
	// DrawCircle(position Position, radius int)
	// DrawLine(start, end Position)
	// DrawText(text string, position Position)
	// DrawImage(image Image, position Position)
}
