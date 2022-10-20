package shape

import "github.com/mokiat/gomath/dprec"

// NewStaticLine creates a new StaticLine shape.
func NewStaticLine(a, b Point) StaticLine {
	return StaticLine{
		a: a,
		b: b,
	}
}

// StaticLine represents an immutable line segment between two 3D points.
type StaticLine struct {
	a Point
	b Point
}

// A returns the starting point of this StaticLine.
func (l StaticLine) A() Point {
	return l.a
}

// B returns the ending point of this StaticLine.
func (l StaticLine) B() Point {
	return l.b
}

// Length returns the length of this StaticLine.
func (l StaticLine) Length() float64 {
	return dprec.Vec3Diff(dprec.Vec3(l.b), dprec.Vec3(l.a)).Length()
}
