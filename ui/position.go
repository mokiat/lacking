package ui

import "fmt"

// NewPosition creates a new Position with the
// specified coordinates.
func NewPosition(x, y int) Position {
	return Position{
		X: x,
		Y: y,
	}
}

// Position represents a position on the screen
// that can either be absolute or relative, depending
// on the context.
type Position struct {
	X int
	Y int
}

// Inverse returns a new Position that is the reverse
// of this Position.
func (p Position) Inverse() Position {
	return Position{
		X: -p.X,
		Y: -p.Y,
	}
}

// Translate returns a new Position that is translated
// by the specified amount.
func (p Position) Translate(dX, dY int) Position {
	return Position{
		X: p.X + dX,
		Y: p.Y + dY,
	}
}

// String returns a string representation of this Position.
func (p Position) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}
