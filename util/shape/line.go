package shape

import "github.com/mokiat/gomath/dprec"

// NewStaticLine creates a new StaticLine.
func NewStaticLine(a, b dprec.Vec3) StaticLine {
	return StaticLine{
		a: a,
		b: b,
	}
}

// StaticLine represents a line segment between two 3D points that cannot
// be resized.
type StaticLine struct {
	a dprec.Vec3
	b dprec.Vec3
}

// A returns the starting point of the line.
func (l StaticLine) A() dprec.Vec3 {
	return l.a
}

// B returns the ending point of the line.
func (l StaticLine) B() dprec.Vec3 {
	return l.b
}

// Length returns the length of the line.
func (l StaticLine) Length() float64 {
	return dprec.Vec3Diff(l.b, l.a).Length()
}
