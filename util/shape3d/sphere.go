package shape3d

import "github.com/mokiat/gomath/dprec"

func NewSphere(position dprec.Vec3, radius float64) Sphere {
	return Sphere{
		Position: position,
		Radius:   radius,
	}
}

type Sphere struct {
	Position dprec.Vec3
	Radius   float64
}
