package shape2d

import "github.com/mokiat/gomath/sprec"

// Square represents an axis-aligned square shape.
type Square struct {
	// Center specifies the center point of the square.
	Center sprec.Vec2
	// Size specifies the length of the square's sides.
	Size float32
}

// ContainsPoint returns whether the specified point lies within the square.
func (s Square) ContainsPoint(point sprec.Vec2) bool {
	halfSize := s.Size * 0.5
	minX := s.Center.X - halfSize
	maxX := s.Center.X + halfSize
	minY := s.Center.Y - halfSize
	maxY := s.Center.Y + halfSize
	return point.X >= minX && point.X <= maxX && point.Y >= minY && point.Y <= maxY
}

// BoundingCircle returns the smallest Circle that fully encompasses the square.
func (s Square) BoundingCircle() Circle {
	halfSize := s.Size * 0.5
	return Circle{
		Center: s.Center,
		Radius: halfSize * sprec.Sqrt(2.0),
	}
}
