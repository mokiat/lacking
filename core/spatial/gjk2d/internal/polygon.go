package internal

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

type Polygon struct {
	Rotation    shape2d.Rotation
	InvRotation shape2d.Rotation
	Points      []dprec.Vec2
}

func (p *Polygon) InitialPoint() dprec.Vec2 {
	return p.Rotation.Apply(p.Points[0])
}

func (p *Polygon) Support(dir dprec.Vec2) dprec.Vec2 {
	dir = p.InvRotation.Apply(dir)
	best := p.Points[0]
	bestDot := dprec.Vec2Dot(best, dir)
	for _, v := range p.Points[1:] {
		if dot := dprec.Vec2Dot(v, dir); dot > bestDot {
			bestDot = dot
			best = v
		}
	}
	return p.Rotation.Apply(best)
}
