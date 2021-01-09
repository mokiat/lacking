package ui

func NewPosition(x, y int) Position {
	return Position{
		X: x,
		Y: y,
	}
}

type Position struct {
	X int
	Y int
}

func (p Position) Inverse() Position {
	return Position{
		X: -p.X,
		Y: -p.Y,
	}
}

func (p Position) Translate(dX, dY int) Position {
	return Position{
		X: p.X + dX,
		Y: p.Y + dY,
	}
}

func NewSize(width, height int) Size {
	return Size{
		Width:  width,
		Height: height,
	}
}

type Size struct {
	Width  int
	Height int
}

func (s Size) Grow(dWidth, dHeight int) Size {
	return Size{
		Width:  s.Width + dWidth,
		Height: s.Height + dHeight,
	}
}

func (s Size) Shrink(dWidth, dHeight int) Size {
	return s.Grow(-dWidth, -dHeight)
}

func (s Size) Empty() bool {
	return s.Width <= 0 || s.Height <= 0
}

type Bounds struct {
	Position
	Size
}

func (b Bounds) Translate(delta Position) Bounds {
	return Bounds{
		Position: b.Position.Translate(delta.X, delta.Y),
		Size:     b.Size,
	}
}

func (b Bounds) Grow(size Size) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     b.Size.Grow(size.Width, size.Height),
	}
}

func (b Bounds) Shrink(size Size) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     b.Size.Shrink(size.Width, size.Height),
	}
}

func (b Bounds) Resize(width, height int) Bounds {
	return Bounds{
		Position: b.Position,
		Size:     NewSize(width, height),
	}
}

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
