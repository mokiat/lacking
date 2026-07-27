package ui

import "fmt"

// Position represents a point on the screen that can either be absolute
// or relative to some parent entity (e.g. an [Element]), depending on the
// context.
//
// The X axis grows to the right and the Y axis grows downwards, meaning
// that the origin is at the top-left corner of the reference area.
type Position struct {
	// X specifies the horizontal coordinate of the position.
	X int
	// Y specifies the vertical coordinate of the position.
	Y int
}

// NewPosition creates a new [Position] with the specified coordinates.
func NewPosition(x, y int) Position {
	return Position{
		X: x,
		Y: y,
	}
}

// Inverse returns a new [Position] that has both of its coordinates
// negated. Translating by the result undoes a [Position.Translate] with
// the original position.
func (p Position) Inverse() Position {
	return Position{
		X: -p.X,
		Y: -p.Y,
	}
}

// Translate returns a new [Position] that is offset by the specified
// delta amount.
func (p Position) Translate(delta Position) Position {
	return Position{
		X: p.X + delta.X,
		Y: p.Y + delta.Y,
	}
}

// String returns a string representation of this [Position] in the
// form "(X, Y)".
func (p Position) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

// Size represents the dimensions of a rectangular area on the screen.
//
// A Size that has a zero or a negative dimension describes an area
// without any content and is considered empty, as reported by
// [Size.Empty].
type Size struct {
	// Width specifies the horizontal dimension of the area.
	Width int
	// Height specifies the vertical dimension of the area.
	Height int
}

// NewSize returns a new [Size] with the specified dimensions.
func NewSize(width, height int) Size {
	return Size{
		Width:  width,
		Height: height,
	}
}

// Inverse returns a new [Size] that has both of its dimensions negated.
// It is mostly useful for turning a growth delta into a shrink one and
// vice versa.
//
// Note that inverting a non-empty [Size] always produces an empty one.
func (s Size) Inverse() Size {
	return Size{
		Width:  -s.Width,
		Height: -s.Height,
	}
}

// Grow returns a new [Size] that is larger than this Size by the given
// delta amount in each direction. A negative delta dimension makes the
// result smaller instead.
func (s Size) Grow(delta Size) Size {
	return Size{
		Width:  s.Width + delta.Width,
		Height: s.Height + delta.Height,
	}
}

// Shrink returns a new [Size] that is smaller than this Size by the given
// delta amount in each direction. A negative delta dimension makes the
// result larger instead.
//
// The dimensions are not clamped, hence the result can end up being empty
// or even negative.
func (s Size) Shrink(delta Size) Size {
	return s.Grow(delta.Inverse())
}

// Empty returns whether this Size describes an area without any content,
// which is the case when either of its dimensions is zero or negative.
func (s Size) Empty() bool {
	return s.Width <= 0 || s.Height <= 0
}

// String returns the string representation of this [Size] in the form
// "(Width, Height)".
func (s Size) String() string {
	return fmt.Sprintf("(%d, %d)", s.Width, s.Height)
}

// Bounds represents a rectangular area on the screen. It is described by
// the [Position] of its top-left corner and its [Size].
//
// The area that is covered includes the top-left corner but excludes the
// bottom-right one. In other words, it spans horizontally from X, which is
// inclusive, to X+Width, which is exclusive, and vertically from Y, which
// is inclusive, to Y+Height, which is exclusive.
type Bounds struct {
	Position
	Size
}

// NewBounds creates a new [Bounds] with a top-left corner at the specified
// coordinates and with the specified dimensions.
func NewBounds(x, y, width, height int) Bounds {
	return Bounds{
		Position: NewPosition(x, y),
		Size:     NewSize(width, height),
	}
}

// Contains returns whether the specified [Position] is contained by these
// Bounds. The top and left edges are inclusive, whereas the bottom and
// right ones are exclusive. An empty [Bounds] contains no positions.
func (b Bounds) Contains(position Position) bool {
	return position.X >= b.X &&
		position.Y >= b.Y &&
		position.X < b.X+b.Width &&
		position.Y < b.Y+b.Height
}

// Translate returns a new [Bounds] that has its position offset by the
// given delta amount and its size preserved.
func (b Bounds) Translate(delta Position) Bounds {
	return Bounds{
		Position: b.Position.Translate(delta),
		Size:     b.Size,
	}
}

// Grow returns a new [Bounds] that has a size that is larger by the given
// amount compared to these Bounds. The top-left corner is preserved, hence
// the area expands to the right and downwards.
func (b Bounds) Grow(size Size) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     b.Size.Grow(size),
	}
}

// Shrink returns a new [Bounds] that has a size that is smaller by the
// given amount compared to these Bounds. The top-left corner is preserved,
// hence the area contracts from the right and the bottom.
//
// The size is not clamped, hence the result can end up being empty.
func (b Bounds) Shrink(size Size) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     b.Size.Shrink(size),
	}
}

// Resize returns a new [Bounds] that has a new [Size] of the specified
// dimensions and an unchanged position.
func (b Bounds) Resize(width, height int) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     NewSize(width, height),
	}
}

// Intersect returns a new [Bounds] that represents the area which is
// covered by both the specified Bounds and these Bounds.
//
// When the two do not overlap, the returned [Bounds] is empty, as reported
// by [Size.Empty], though its exact position and (negative) dimensions are
// unspecified. Users should check [Size.Empty] on the result before making
// use of its position or dimensions.
func (b Bounds) Intersect(other Bounds) Bounds {
	position := NewPosition(
		max(b.X, other.X),
		max(b.Y, other.Y),
	)
	size := NewSize(
		min(b.X+b.Width, other.X+other.Width)-position.X,
		min(b.Y+b.Height, other.Y+other.Height)-position.Y,
	)
	return Bounds{
		Position: position,
		Size:     size,
	}
}

// String returns the string representation of these [Bounds] in the form
// "((X, Y), (Width, Height))".
func (b Bounds) String() string {
	return fmt.Sprintf("(%s, %s)", b.Position, b.Size)
}
