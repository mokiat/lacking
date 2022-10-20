package shape

// NewStaticSphere creates a new StaticSphere shape.
func NewStaticSphere(radius float64) StaticSphere {
	return StaticSphere{
		Transform: IdentityTransform(),
		radius:    radius,
	}
}

// StaticSphere represents an immutable sphere shape.
type StaticSphere struct {
	Transform
	radius float64
}

// BoundingSphereRadius returns the radius of a sphere that can encompass
// this shape.
func (s StaticSphere) BoundingSphereRadius() float64 {
	return s.radius
}

// Radius returns the radius of this StaticSphere.
func (s StaticSphere) Radius() float64 {
	return s.radius
}

// Diameter returns the diameter of this StaticSphere.
func (s StaticSphere) Diameter() float64 {
	return s.radius * 2.0
}

// WithRadius returns a new StaticSphere that is based on this one but has the
// specified radius.
func (s StaticSphere) WithRadius(radius float64) StaticSphere {
	s.radius = radius
	return s
}

// WithTransform returns a new StaticSphere that is based on this one but has
// the specified transform.
func (s StaticSphere) WithTransform(transform Transform) StaticSphere {
	s.Transform = transform
	return s
}

// Transformed returns a new StaticSphere that is based on this one but has
// the specified transform applied to it.
func (s StaticSphere) Transformed(parent Transform) StaticSphere {
	s.Transform = s.Transform.Transformed(parent)
	return s
}
