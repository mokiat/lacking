package gjk2d

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

type Shape struct {
	Points     []sprec.Vec2
	SkinRadius float32
}

func ShapeFromCircle(circle shape2d.Circle) Shape {
	return Shape{
		Points:     []sprec.Vec2{circle.Center},
		SkinRadius: circle.Radius,
	}
}

func ShapeFromSquare(square shape2d.Square) Shape {
	halfSize := square.Size / 2.0
	return Shape{
		Points: []sprec.Vec2{
			sprec.Vec2Sum(square.Center, sprec.NewVec2(-halfSize, -halfSize)),
			sprec.Vec2Sum(square.Center, sprec.NewVec2(halfSize, -halfSize)),
			sprec.Vec2Sum(square.Center, sprec.NewVec2(halfSize, halfSize)),
			sprec.Vec2Sum(square.Center, sprec.NewVec2(-halfSize, halfSize)),
		},
		SkinRadius: 0.0,
	}
}

func ShapeFromRoundSquare(square shape2d.Square, radius float32) Shape {
	halfSize := square.Size / 2.0
	return Shape{
		Points: []sprec.Vec2{
			sprec.Vec2Sum(square.Center, sprec.NewVec2(-halfSize, -halfSize)),
			sprec.Vec2Sum(square.Center, sprec.NewVec2(halfSize, -halfSize)),
			sprec.Vec2Sum(square.Center, sprec.NewVec2(halfSize, halfSize)),
			sprec.Vec2Sum(square.Center, sprec.NewVec2(-halfSize, halfSize)),
		},
		SkinRadius: radius,
	}
}
