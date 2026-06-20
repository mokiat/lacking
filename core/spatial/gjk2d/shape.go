package gjk2d

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

type Rotation = shape2d.Rotation

type Shape struct {
	Position   sprec.Vec2
	Rotation   Rotation
	Points     []sprec.Vec2
	SkinRadius float32
}

func ShapeFromCircle(circle shape2d.Circle) Shape {
	return Shape{
		Position: circle.Center,
		Rotation: shape2d.IdentityRotation(),
		Points: []sprec.Vec2{
			sprec.ZeroVec2(),
		},
		SkinRadius: circle.Radius,
	}
}

func ShapeFromCapsule(capsule shape2d.Capsule) Shape {
	return Shape{
		Position: sprec.ZeroVec2(),
		Rotation: shape2d.IdentityRotation(),
		Points: []sprec.Vec2{
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
		Points: []sprec.Vec2{
			sprec.NewVec2(-halfSize, -halfSize),
			sprec.NewVec2(halfSize, -halfSize),
			sprec.NewVec2(halfSize, halfSize),
			sprec.NewVec2(-halfSize, halfSize),
		},
		SkinRadius: 0.0,
	}
}

func ShapeFromSquareRound(square shape2d.Square, radius float32) Shape {
	halfSize := square.Size / 2.0
	return Shape{
		Position: square.Center,
		Rotation: square.Rotation,
		Points: []sprec.Vec2{
			sprec.NewVec2(-halfSize, -halfSize),
			sprec.NewVec2(halfSize, -halfSize),
			sprec.NewVec2(halfSize, halfSize),
			sprec.NewVec2(-halfSize, halfSize),
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
		Points: []sprec.Vec2{
			sprec.NewVec2(-halfWidth, -halfHeight),
			sprec.NewVec2(halfWidth, -halfHeight),
			sprec.NewVec2(halfWidth, halfHeight),
			sprec.NewVec2(-halfWidth, halfHeight),
		},
		SkinRadius: 0.0,
	}
}

func ShapeFromRectangleRound(rectangle shape2d.Rectangle, radius float32) Shape {
	halfWidth := rectangle.Width / 2.0
	halfHeight := rectangle.Height / 2.0
	return Shape{
		Position: rectangle.Center,
		Rotation: rectangle.Rotation,
		Points: []sprec.Vec2{
			sprec.NewVec2(-halfWidth, -halfHeight),
			sprec.NewVec2(halfWidth, -halfHeight),
			sprec.NewVec2(halfWidth, halfHeight),
			sprec.NewVec2(-halfWidth, halfHeight),
		},
		SkinRadius: radius,
	}
}
