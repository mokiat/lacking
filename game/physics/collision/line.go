package collision

import "github.com/mokiat/gomath/dprec"

// NewLine creates a new Line shape.
func NewLine(a, b dprec.Vec3) Line {
	return Line{
		a: a,
		b: b,
	}
}

// Line represents a line segment between two 3D points.
type Line struct {
	a dprec.Vec3
	b dprec.Vec3
}

// Replace replaces this shape with the template one after the specified
// transformation has been applied to it.
func (l *Line) Replace(template Line, transform Transform) {
	l.a = transform.Vector(template.a)
	l.b = transform.Vector(template.b)
}

// A returns the starting point of this line.
func (l *Line) A() dprec.Vec3 {
	return l.a
}

// B returns the ending point of this line.
func (l *Line) B() dprec.Vec3 {
	return l.b
}

// Length returns the length of this line.
func (l *Line) Length() float64 {
	return dprec.Vec3Diff(l.b, l.a).Length()
}
