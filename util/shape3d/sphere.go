package shape3d

import "github.com/mokiat/gomath/dprec"

// NewSphere creates a new sphere with the specified position and radius.
func NewSphere(position dprec.Vec3, radius float64) Sphere {
	return Sphere{
		Position: position,
		Radius:   radius,
	}
}

// TransformedSphere creates a new sphere based off of an existing one
// by applying the specified transformation.
func TransformedSphere(source Sphere, transform Transform) Sphere {
	return Sphere{
		Position: transform.Apply(source.Position),
		Radius:   source.Radius,
	}
}

// Sphere represents a sphere shape.
type Sphere struct {

	// Position is the position of the sphere.
	Position dprec.Vec3

	// Radius is the radius of the sphere.
	Radius float64
}
