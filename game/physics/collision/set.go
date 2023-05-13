package collision

import "github.com/mokiat/gomath/dprec"

// SetOption represents a configuration option for a Set object.
type SetOption func(s *Set)

// WithSpheres specifies the sphere shapes contained by the set.
func WithSpheres(spheres []Sphere) SetOption {
	return func(s *Set) {
		s.spheres = spheres
	}
}

// WithBoxes specifies the box shapes contained by the set.
func WithBoxes(boxes []Box) SetOption {
	return func(s *Set) {
		s.boxes = boxes
	}
}

// WithMeshes specifies the mesh shapes contained by the set.
func WithMeshes(meshes []Mesh) SetOption {
	return func(s *Set) {
		s.meshes = meshes
	}
}

// NewSet constructs a new collision Set with the specified options.
func NewSet(opts ...SetOption) Set {
	result := Set{}
	for _, opt := range opts {
		opt(&result)
	}
	center := result.calculateCenter()
	radius := result.calculateRadius(center)
	result.bs = NewSphere(center, radius)
	return result
}

// Set represents a collection of collision shapes.
type Set struct {
	spheres []Sphere
	boxes   []Box
	meshes  []Mesh
	bs      Sphere
}

// Replace replaces this shape with the template one after the specified
// transformation has been applied to it.
func (s *Set) Replace(template Set, transform Transform) {
	if len(s.spheres) != len(template.spheres) {
		s.spheres = make([]Sphere, len(template.spheres))
	}
	for i := range s.spheres {
		s.spheres[i].Replace(template.spheres[i], transform)
	}

	if len(s.boxes) != len(template.boxes) {
		s.boxes = make([]Box, len(template.boxes))
	}
	for i := range s.boxes {
		s.boxes[i].Replace(template.boxes[i], transform)
	}

	if len(s.meshes) != len(template.meshes) {
		s.meshes = make([]Mesh, len(template.meshes))
	}
	for i := range s.meshes {
		s.meshes[i].Replace(template.meshes[i], transform)
	}

	s.bs.Replace(template.bs, transform)
}

// IsEmpty returns whether this set is shapeless.
func (s *Set) IsEmpty() bool {
	return len(s.spheres) == 0 && len(s.boxes) == 0 && len(s.meshes) == 0
}

// Spheres returns all spheres contained in this set.
//
// NOTE: This returns the internal slice which should not be modified
// in any way.
func (s *Set) Spheres() []Sphere {
	return s.spheres
}

// Boxes returns all boxes contained in this set.
//
// NOTE: This returns the internal slice which should not be modified
// in any way.
func (s *Set) Boxes() []Box {
	return s.boxes
}

// Meshes returns all meshes contained in this set.
//
// NOTE: This returns the internal slice which should not be modified
// in any way.
func (s *Set) Meshes() []Mesh {
	return s.meshes
}

// BoundingSphere returns a sphere that encompases this set.
func (s *Set) BoundingSphere() Sphere {
	return s.bs
}

func (s *Set) calculateCenter() dprec.Vec3 {
	totalCount := len(s.spheres) + len(s.boxes) + len(s.meshes)
	if totalCount == 0 {
		return dprec.ZeroVec3()
	}
	var center dprec.Vec3
	for _, sphere := range s.spheres {
		bs := sphere
		center = dprec.Vec3Sum(center, bs.Position())
	}
	for _, box := range s.boxes {
		bs := box.BoundingSphere()
		center = dprec.Vec3Sum(center, bs.Position())
	}
	for _, mesh := range s.meshes {
		bs := mesh.BoundingSphere()
		center = dprec.Vec3Sum(center, bs.Position())
	}
	return dprec.Vec3Quot(center, float64(totalCount))
}

func (s *Set) calculateRadius(from dprec.Vec3) float64 {
	var radius float64
	for _, sphere := range s.spheres {
		bs := sphere
		radius = dprec.Max(radius, dprec.Vec3Diff(bs.Position(), from).Length()+bs.Radius())
	}
	for _, box := range s.boxes {
		bs := box.BoundingSphere()
		radius = dprec.Max(radius, dprec.Vec3Diff(bs.Position(), from).Length()+bs.Radius())
	}
	for _, mesh := range s.meshes {
		bs := mesh.BoundingSphere()
		radius = dprec.Max(radius, dprec.Vec3Diff(bs.Position(), from).Length()+bs.Radius())
	}
	return radius
}
