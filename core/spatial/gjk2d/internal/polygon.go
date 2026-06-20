package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

type Polygon struct {
	Rotation    shape2d.Rotation
	InvRotation shape2d.Rotation
	Points      []sprec.Vec2
}

func (p *Polygon) InitialPoint() sprec.Vec2 {
	return p.Rotation.Apply(p.Points[0])
}

func (p *Polygon) Support(dir sprec.Vec2) sprec.Vec2 {
	dir = p.InvRotation.Apply(dir)
	best := p.Points[0]
	bestDot := sprec.Vec2Dot(best, dir)
	for _, v := range p.Points[1:] {
		if dot := sprec.Vec2Dot(v, dir); dot > bestDot {
			bestDot = dot
			best = v
		}
	}
	return p.Rotation.Apply(best)
}
