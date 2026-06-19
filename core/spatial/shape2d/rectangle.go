package shape2d

import "github.com/mokiat/gomath/sprec"

// Rectangle represents an axis-aligned rectangle shape.
type Rectangle struct {
	// Center specifies the center point of the rectangle.
	Center sprec.Vec2
	// Width specifies the length of the rectangle's horizontal sides.
	Width float32
	// Height specifies the length of the rectangle's vertical sides.
	Height float32
}

// ContainsPoint returns whether the specified point lies within the rectangle.
func (r Rectangle) ContainsPoint(point sprec.Vec2) bool {
	halfWidth := r.Width * 0.5
	halfHeight := r.Height * 0.5
	minX := r.Center.X - halfWidth
	maxX := r.Center.X + halfWidth
	minY := r.Center.Y - halfHeight
	maxY := r.Center.Y + halfHeight
	return point.X >= minX && point.X <= maxX && point.Y >= minY && point.Y <= maxY
}

// BoundingCircle returns the smallest Circle that fully encompasses the rectangle.
func (r Rectangle) BoundingCircle() Circle {
	halfWidth := r.Width * 0.5
	halfHeight := r.Height * 0.5
	return Circle{
		Center: r.Center,
		Radius: sprec.Sqrt(halfWidth*halfWidth + halfHeight*halfHeight),
	}
}
