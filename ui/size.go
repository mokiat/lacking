package ui

import "fmt"

// NewSize returns a new Size with the specified dimensions.
func NewSize(width, height int) Size {
	return Size{
		Width:  width,
		Height: height,
	}
}

// Size represents the dimensions of something on the screen.
type Size struct {
	Width  int
	Height int
}

// Inverse returns the inverse Size of the current one.
func (s Size) Inverse() Size {
	return Size{
		Width:  -s.Width,
		Height: -s.Height,
	}
}

// Grow returns a new Size that is larger than this Size
// by the given delta amount.
func (s Size) Grow(delta Size) Size {
	return Size{
		Width:  s.Width + delta.Width,
		Height: s.Height + delta.Height,
	}
}

// Shrink returns a new Size that is smaller than this Size
// by the given delta amount.
func (s Size) Shrink(delta Size) Size {
	return s.Grow(delta.Inverse())
}

// Empty returns whether this Size is zero or negative
// in any direction.
func (s Size) Empty() bool {
	return s.Width <= 0 || s.Height <= 0
}

// String returns the string representation of this Size.
func (s Size) String() string {
	return fmt.Sprintf("(%d, %d)", s.Width, s.Height)
}
