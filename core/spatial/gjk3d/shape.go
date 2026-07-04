package gjk3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// Rotation is an alias for the 3D rotation type [shape3d.Rotation].
type Rotation = shape3d.Rotation

// Shape represents a convex polyhedron with an optional skin radius, used as
// input to the GJK algorithm. The skin radius extends the effective boundary
// of the shape outward, enabling rounded edges, spheres, and capsules.
type Shape struct {
	// Position is the world-space origin of the shape.
	Position dprec.Vec3
	// Rotation describes the orientation of the shape's local axes in world space.
	Rotation Rotation
	// Points holds the local-space vertices of the convex polyhedron core,
	// before rotation is applied. An empty slice is treated as a degenerate
	// shape that does not intersect anything.
	Points []dprec.Vec3
	// SkinRadius extends the effective boundary of the shape outward by this
	// distance. A value of zero means no extension.
	SkinRadius float64
}

// ShapeFromSegment constructs a [Shape] from a [shape3d.Segment]. The segment is
// represented as a two-point polyhedron with no skin radius.
func ShapeFromSegment(segment shape3d.Segment) Shape {
	return Shape{
		Position: dprec.ZeroVec3(),
		Rotation: shape3d.IdentityRotation(),
		Points: []dprec.Vec3{
			segment.A,
			segment.B,
		},
		SkinRadius: 0.0,
	}
}

// ShapeFromSphere constructs a [Shape] from a [shape3d.Sphere]. The sphere is
// represented as a single central point with a skin radius equal to the
// sphere's radius.
func ShapeFromSphere(sphere shape3d.Sphere) Shape {
	return Shape{
		Position: sphere.Center,
		Rotation: shape3d.IdentityRotation(),
		Points: []dprec.Vec3{
			dprec.ZeroVec3(),
		},
		SkinRadius: sphere.Radius,
	}
}

// ShapeFromCapsule constructs a [Shape] from a [shape3d.Segment] and a
// radius. The capsule is represented as a two-point segment with a skin
// radius equal to the capsule's radius.
func ShapeFromCapsule(segment shape3d.Segment, radius float64) Shape {
	return Shape{
		Position: dprec.ZeroVec3(),
		Rotation: shape3d.IdentityRotation(),
		Points: []dprec.Vec3{
			segment.A,
			segment.B,
		},
		SkinRadius: radius,
	}
}

// ShapeFromTriangle constructs a [Shape] from a [shape3d.Triangle] with no
// skin radius.
func ShapeFromTriangle(triangle shape3d.Triangle) Shape {
	return Shape{
		Position: dprec.ZeroVec3(),
		Rotation: shape3d.IdentityRotation(),
		Points: []dprec.Vec3{
			triangle.A,
			triangle.B,
			triangle.C,
		},
		SkinRadius: 0.0,
	}
}

// ShapeFromBox constructs a [Shape] from a [shape3d.Box] with no skin radius.
func ShapeFromBox(box shape3d.Box) Shape {
	return Shape{
		Position:   box.Center,
		Rotation:   box.Rotation,
		Points:     boxPoints(box),
		SkinRadius: 0.0,
	}
}

// ShapeFromBoxRound constructs a [Shape] from a [shape3d.Box] with the given
// skin radius, producing a rounded box.
func ShapeFromBoxRound(box shape3d.Box, radius float64) Shape {
	return Shape{
		Position:   box.Center,
		Rotation:   box.Rotation,
		Points:     boxPoints(box),
		SkinRadius: radius,
	}
}

// WSPosition returns the world-space position of the polyhedron vertex at
// localIndex, taking the shape's position and rotation into account.
func (s *Shape) WSPosition(localIndex int) dprec.Vec3 {
	return dprec.Vec3Sum(s.Position, s.Rotation.Apply(s.Points[localIndex]))
}

// boxPoints returns the eight local-space corner vertices of the box.
func boxPoints(box shape3d.Box) []dprec.Vec3 {
	halfWidth := box.HalfWidth
	halfHeight := box.HalfHeight
	halfLength := box.HalfLength
	return []dprec.Vec3{
		dprec.NewVec3(-halfWidth, -halfHeight, -halfLength),
		dprec.NewVec3(halfWidth, -halfHeight, -halfLength),
		dprec.NewVec3(halfWidth, halfHeight, -halfLength),
		dprec.NewVec3(-halfWidth, halfHeight, -halfLength),
		dprec.NewVec3(-halfWidth, -halfHeight, halfLength),
		dprec.NewVec3(halfWidth, -halfHeight, halfLength),
		dprec.NewVec3(halfWidth, halfHeight, halfLength),
		dprec.NewVec3(-halfWidth, halfHeight, halfLength),
	}
}
