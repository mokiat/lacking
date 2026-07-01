package internal

import "github.com/mokiat/gomath/dprec"

// originProjectsToEdge reports whether the perpendicular projection of the
// origin onto the line through start and end falls within the segment.
func originProjectsToEdge(start, end dprec.Vec2) bool {
	edge := dprec.Vec2Diff(end, start)
	dot := -dprec.Vec2Dot(edge, start)
	return dot >= 0.0 && dot <= edge.SqrLength()
}

// transposeVec2 returns v rotated by 90 degrees clockwise. For a directed
// edge this is the right-hand normal.
func transposeVec2(v dprec.Vec2) dprec.Vec2 {
	return dprec.NewVec2(v.Y, -v.X)
}
