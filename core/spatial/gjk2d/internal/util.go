package internal

import "github.com/mokiat/gomath/dprec"

// transposeVec2 returns v rotated by 90 degrees clockwise. For a directed
// edge this is the right-hand normal.
func transposeVec2(v dprec.Vec2) dprec.Vec2 {
	return dprec.NewVec2(v.Y, -v.X)
}
