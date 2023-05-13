package collision

import "github.com/mokiat/gomath/dprec"

// NewSphere creates a new Sphere shape.
func NewSphere(position dprec.Vec3, radius float64) Sphere {
	return Sphere{
		position: position,
		radius:   radius,
	}
}

// Sphere represents a 3D sphere shape.
type Sphere struct {
	position dprec.Vec3
	radius   float64
}

// Replace replaces this shape with the template one after the specified
// transformation has been applied to it.
func (s *Sphere) Replace(template Sphere, transform Transform) {
	s.position = transform.Vector(template.position)
	s.radius = template.radius
}

// Position returns the location of this sphere.
func (s *Sphere) Position() dprec.Vec3 {
	return s.position
}

// Radius returns the radius of this sphere.
func (s *Sphere) Radius() float64 {
	return s.radius
}

// Diameter returns the diameter of this sphere.
func (s *Sphere) Diameter() float64 {
	return s.radius * 2.0
}
