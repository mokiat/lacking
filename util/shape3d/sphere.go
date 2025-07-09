package shape3d

import (
	"github.com/mokiat/gomath/dprec"
)

func NewSphere(position dprec.Vec3, radius float64) Sphere {
	return Sphere{
		Position: position,
		Radius:   radius,
	}
}

func TransformedSphere(source Sphere, transform Transform) Sphere {
	return Sphere{
		Position: transform.Apply(source.Position),
		Radius:   source.Radius,
	}
}

type Sphere struct {
	Position dprec.Vec3
	Radius   float64
}
