package ui

import "fmt"

// NewBounds creates a new Bounds object.
func NewBounds(x, y, width, height int) Bounds {
	return Bounds{
		Position: NewPosition(x, y),
		Size:     NewSize(width, height),
	}
}

// Bounds represents a content area on the screen. It
// consists of a Position and Size.
type Bounds struct {
	Position
	Size
}

// Contains returns whether the specified Position is
// contained by this Bounds.
func (b Bounds) Contains(position Position) bool {
	return position.X >= b.X &&
		position.Y >= b.Y &&
		position.X < b.X+b.Width &&
		position.Y < b.Y+b.Height
}

// Translate returns a new Bounds that is with a translated
// position by the given amount.
func (b Bounds) Translate(delta Position) Bounds {
	return Bounds{
		Position: b.Position.Translate(delta),
		Size:     b.Size,
	}
}

// Grow returns a new Bounds that has a size that is
// larger by the given amount compared to these Bounds.
func (b Bounds) Grow(size Size) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     b.Size.Grow(size),
	}
}

// Shrink returns a new Bounds that has a size that is
//smaller by the given amount compared to these Bounds.
func (b Bounds) Shrink(size Size) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     b.Size.Shrink(size),
	}
}

// Resize returns a new Bounds that is with a new Size
// of the specified dimensions.
func (b Bounds) Resize(width, height int) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     NewSize(width, height),
	}
}

// Intersect returns a new Bounds that is the intersection
// of the specified Bounds and these Bounds.
func (b Bounds) Intersect(other Bounds) Bounds {
	position := NewPosition(
		maxInt(b.X, other.X),
		maxInt(b.Y, other.Y),
	)
	size := NewSize(
		minInt(b.X+b.Width, other.X+other.Width)-position.X,
		minInt(b.Y+b.Height, other.Y+other.Height)-position.Y,
	)
	return Bounds{
		Position: position,
		Size:     size,
	}
}

// String returns the string representation of these Bounds.
func (b Bounds) String() string {
	return fmt.Sprintf("(%s, %s)", b.Position, b.Size)
}
