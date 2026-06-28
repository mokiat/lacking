package shape3d

import "github.com/mokiat/gomath/dprec"

// Sphere represents a three-dimensional sphere shape.
type Sphere struct {
	// Center specifies the center point of the sphere.
	Center dprec.Vec3
	// Radius specifies the radius of the sphere.
	Radius float64
}

// TransformedSphere returns a new [Sphere] that is the result of applying the
// specified transform to the given sphere. The center is moved by the transform
// while the radius is left unchanged, since a rigid-body transform preserves
// distances.
func TransformedSphere(sphere Sphere, transform Transform) Sphere {
	return Sphere{
		Center: transform.Apply(sphere.Center),
		Radius: sphere.Radius,
	}
}

// ContainsPoint returns whether the specified point lies within the sphere.
func (s Sphere) ContainsPoint(point dprec.Vec3) bool {
	delta := dprec.Vec3Diff(point, s.Center)
	return delta.SqrLength() <= s.Radius*s.Radius
}
