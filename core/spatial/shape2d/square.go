package shape2d

import "github.com/mokiat/gomath/dprec"

// Square represents a two-dimensional square shape.
type Square struct {
	// Center specifies the center point of the square.
	Center dprec.Vec2
	// Rotation specifies the orientation of the square.
	Rotation Rotation
	// Size specifies the length of the square's sides.
	Size float64
}

// ContainsPoint returns whether the specified point lies within the square.
func (s Square) ContainsPoint(point dprec.Vec2) bool {
	offset := dprec.Vec2Diff(point, s.Center)
	localPoint := s.Rotation.Inverse().Apply(offset)
	halfSize := s.Size * 0.5
	return localPoint.X >= -halfSize &&
		localPoint.X <= halfSize &&
		localPoint.Y >= -halfSize &&
		localPoint.Y <= halfSize
}

// BoundingCircle returns the smallest [Circle] that fully encompasses the square.
func (s Square) BoundingCircle() Circle {
	halfSize := s.Size * 0.5
	return Circle{
		Center: s.Center,
		Radius: halfSize * dprec.Sqrt(2.0),
	}
}
