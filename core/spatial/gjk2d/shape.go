package gjk2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

type Rotation = shape2d.Rotation

type Shape struct {
	Position   dprec.Vec2
	Rotation   Rotation
	Points     []dprec.Vec2
	SkinRadius float64
}

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

func ShapeFromSquare(square shape2d.Square) Shape {
	halfSize := square.Size / 2.0
	return Shape{
		Position: square.Center,
		Rotation: square.Rotation,
		Points: []dprec.Vec2{
			dprec.NewVec2(-halfSize, -halfSize),
			dprec.NewVec2(halfSize, -halfSize),
			dprec.NewVec2(halfSize, halfSize),
			dprec.NewVec2(-halfSize, halfSize),
		},
		SkinRadius: 0.0,
	}
}

func ShapeFromSquareRound(square shape2d.Square, radius float64) Shape {
	halfSize := square.Size / 2.0
	return Shape{
		Position: square.Center,
		Rotation: square.Rotation,
		Points: []dprec.Vec2{
			dprec.NewVec2(-halfSize, -halfSize),
			dprec.NewVec2(halfSize, -halfSize),
			dprec.NewVec2(halfSize, halfSize),
			dprec.NewVec2(-halfSize, halfSize),
		},
		SkinRadius: radius,
	}
}

func ShapeFromRectangle(rectangle shape2d.Rectangle) Shape {
	halfWidth := rectangle.Width / 2.0
	halfHeight := rectangle.Height / 2.0
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

func ShapeFromRectangleRound(rectangle shape2d.Rectangle, radius float64) Shape {
	halfWidth := rectangle.Width / 2.0
	halfHeight := rectangle.Height / 2.0
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
