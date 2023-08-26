package ui

import "fmt"

// SymmetricSpacing returns a Spacing that is horizontally and
// vertically symmetric.
func SymmetricSpacing(horizontal, vertical int) Spacing {
	return Spacing{
		Left:   horizontal,
		Right:  horizontal,
		Top:    vertical,
		Bottom: vertical,
	}
}

// UniformSpacing returns a Spacing that is equal all around.
func UniformSpacing(spacing int) Spacing {
	return Spacing{
		Left:   spacing,
		Right:  spacing,
		Top:    spacing,
		Bottom: spacing,
	}
}

// Spacing represents a spacing around or inside a given
// screen entity (e.g. Element).
type Spacing struct {
	Left   int
	Right  int
	Top    int
	Bottom int
}

// Vertical returns the vertical amount of spacing.
func (s Spacing) Vertical() int {
	return s.Top + s.Bottom
}

// Horizontal returns the horizontal amount of spacing.
func (s Spacing) Horizontal() int {
	return s.Left + s.Right
}

// Size returns the amount of spacing used in both horizontal and
// vertical direction.
func (s Spacing) Size() Size {
	return Size{
		Width:  s.Horizontal(),
		Height: s.Vertical(),
	}
}

// String returns the strings representation of this Spacing.
func (s Spacing) String() string {
	return fmt.Sprintf("(%d, %d, %d, %d)", s.Left, s.Right, s.Top, s.Bottom)
}
