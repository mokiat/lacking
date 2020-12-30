package ui

import "github.com/mokiat/gomath/sprec"

type Bounds struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

func (b Bounds) Position() sprec.Vec2 {
	return sprec.NewVec2(b.X, b.Y)
}

func (b Bounds) Size() sprec.Vec2 {
	return sprec.NewVec2(b.Width, b.Height)
}

func (b Bounds) Intersects(other Bounds) bool {
	if other.X >= b.X+b.Width {
		return false
	}
	if other.X+other.Width <= b.X {
		return false
	}
	if other.Y >= b.Y+b.Height {
		return false
	}
	if other.Y+other.Height <= b.Y {
		return false
	}
	return true
}
