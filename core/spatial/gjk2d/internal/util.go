package internal

import "github.com/mokiat/gomath/dprec"

func originProjectsToEdge(start, end dprec.Vec2) bool {
	edge := dprec.Vec2Diff(end, start)
	dot := -dprec.Vec2Dot(edge, start)
	return dot >= 0.0 && dot <= edge.SqrLength()
}

func transposeVec2(v dprec.Vec2) dprec.Vec2 {
	return dprec.NewVec2(v.Y, -v.X)
}
