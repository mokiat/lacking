package shape2d

import "github.com/mokiat/gomath/dprec"

func NewCircle(position dprec.Vec2, radius float64) Circle {
	return Circle{
		Position: position,
		Radius:   radius,
	}
}

type Circle struct {
	Position dprec.Vec2
	Radius   float64
}
