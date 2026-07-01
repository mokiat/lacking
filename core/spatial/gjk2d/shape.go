package gjk2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// Rotation is an alias for the 2D rotation type from the shape2d package.
type Rotation = shape2d.Rotation

// Shape represents a convex polygon with an optional skin radius, used as
// input to the GJK algorithm. The skin radius extends the effective boundary
// of the shape outward, enabling rounded corners, circles, and capsules.
type Shape struct {
	// Position is the world-space origin of the shape.
	Position dprec.Vec2
	// Rotation describes the orientation of the shape's local axes in world space.
	Rotation Rotation
	// Points holds the local-space vertices of the convex polygon core,
	// before rotation is applied. An empty slice is treated as a degenerate
	// shape that does not intersect anything.
	Points []dprec.Vec2
	// SkinRadius extends the effective boundary of the shape outward by this
	// distance. A value of zero means no extension.
	SkinRadius float64
}

// ShapeFromSegment constructs a [Shape] from a [shape2d.Segment]. The segment is
// represented as a two-point polygon with no skin radius.
func ShapeFromSegment(segment shape2d.Segment) Shape {
	return Shape{
		Position: dprec.ZeroVec2(),
		Rotation: shape2d.IdentityRotation(),
		Points: []dprec.Vec2{
			segment.A,
			segment.B,
		},
		SkinRadius: 0.0,
	}
}

// ShapeFromCircle constructs a [Shape] from a [shape2d.Circle]. The circle is represented
// as a single central point with a skin radius equal to the circle's radius.
func ShapeFromCircle(circle shape2d.Circle) Shape {
	return Shape{
		Position: circle.Center,
		Rotation: shape2d.IdentityRotation(),
		Points: []dprec.Vec2{
			dprec.ZeroVec2(),
		},
		SkinRadius: circle.Radius,
	}
}

// ShapeFromCapsule constructs a [Shape] from a [shape2d.Capsule]. The capsule is
// represented as a two-point segment with a skin radius equal to the
// capsule's radius.
func ShapeFromCapsule(capsule shape2d.Capsule) Shape {
	return Shape{
		Position: dprec.ZeroVec2(),
		Rotation: shape2d.IdentityRotation(),
		Points: []dprec.Vec2{
			capsule.A,
			capsule.B,
		},
		SkinRadius: capsule.Radius,
	}
}

// ShapeFromRectangle constructs a [Shape] from a [shape2d.Rectangle] with no skin radius.
func ShapeFromRectangle(rectangle shape2d.Rectangle) Shape {
	halfWidth := rectangle.HalfWidth
	halfHeight := rectangle.HalfHeight
	return Shape{
		Position: rectangle.Center,
		Rotation: rectangle.Rotation,
		Points: []dprec.Vec2{
			dprec.NewVec2(-halfWidth, -halfHeight),
			dprec.NewVec2(halfWidth, -halfHeight),
			dprec.NewVec2(halfWidth, halfHeight),
			dprec.NewVec2(-halfWidth, halfHeight),
		},
		SkinRadius: 0.0,
	}
}

// ShapeFromRectangleRound constructs a [Shape] from a [shape2d.Rectangle] with the given
// skin radius, producing a rounded rectangle.
func ShapeFromRectangleRound(rectangle shape2d.Rectangle, radius float64) Shape {
	halfWidth := rectangle.HalfWidth
	halfHeight := rectangle.HalfHeight
	return Shape{
		Position: rectangle.Center,
		Rotation: rectangle.Rotation,
		Points: []dprec.Vec2{
			dprec.NewVec2(-halfWidth, -halfHeight),
			dprec.NewVec2(halfWidth, -halfHeight),
			dprec.NewVec2(halfWidth, halfHeight),
			dprec.NewVec2(-halfWidth, halfHeight),
		},
		SkinRadius: radius,
	}
}
