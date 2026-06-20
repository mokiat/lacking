package shape2d

import "github.com/mokiat/gomath/sprec"

// Square represents a two-dimensional square shape.
type Square struct {
	// Center specifies the center point of the square.
	Center sprec.Vec2
	// Rotation specifies the orientation of the square.
	Rotation Rotation
	// Size specifies the length of the square's sides.
	Size float32
}

// ContainsPoint returns whether the specified point lies within the square.
func (s Square) ContainsPoint(point sprec.Vec2) bool {
	offset := sprec.Vec2Diff(point, s.Center)
	localPoint := s.Rotation.Inverse().Apply(offset)
	halfSize := s.Size * 0.5
	return localPoint.X >= -halfSize &&
		localPoint.X <= halfSize &&
		localPoint.Y >= -halfSize &&
		localPoint.Y <= halfSize
}

// BoundingCircle returns the smallest Circle that fully encompasses the square.
func (s Square) BoundingCircle() Circle {
	halfSize := s.Size * 0.5
	return Circle{
		Center: s.Center,
		Radius: halfSize * sprec.Sqrt(2.0),
	}
}
