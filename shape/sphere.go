package shape

// NewStaticSphere returns a new StaticSphere.
func NewStaticSphere(radius float64) StaticSphere {
	return StaticSphere{
		radius: radius,
	}
}

// StaticSphere represents a 3D sphere that cannot be resized.
type StaticSphere struct {
	radius float64
}

// BoundingSphereRadius returns the radius of a sphere that can encompass
// this shape.
func (s StaticSphere) BoundingSphereRadius() float64 {
	return s.radius
}

// Radius returns the radius of the sphere.
func (s StaticSphere) Radius() float64 {
	return s.radius
}
