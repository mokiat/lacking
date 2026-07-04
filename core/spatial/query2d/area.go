package query2d

import "github.com/mokiat/lacking/core/spatial/shape2d"

// Area represents the spatial area of an object in the 2D space.
type Area struct {
	x float64
	y float64
	r float64
}

// AreaFromCircle creates an [Area] that covers the given circle.
func AreaFromCircle(circle shape2d.Circle) Area {
	return Area{
		x: circle.Center.X,
		y: circle.Center.Y,
		r: circle.Radius,
	}
}

// AreaFromRectangle creates an [Area] that covers the given rectangle.
//
// The area is a circle centered on the rectangle, with a radius equal to the
// larger of the rectangle's half-width and half-height. This covers the
// rectangle along its shorter axis but not necessarily its corners; it is a
// conservative bound for broad-phase queries rather than an exact fit.
func AreaFromRectangle(rect shape2d.Rectangle) Area {
	return Area{
		x: rect.Center.X,
		y: rect.Center.Y,
		r: max(rect.HalfWidth, rect.HalfHeight),
	}
}
