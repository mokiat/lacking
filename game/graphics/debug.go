package graphics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
)

type Debug struct {
	renderer *sceneRenderer
}

func (d *Debug) Reset() {
	d.renderer.ResetDebugLines()
}

func (d *Debug) Line(start, end, color dprec.Vec3) {
	d.renderer.QueueDebugLine(debugLine{
		Start: dtos.Vec3(start),
		End:   dtos.Vec3(end),
		Color: dtos.Vec3(color),
	})
}

func (d *Debug) Triangle(p1, p2, p3, color dprec.Vec3) {
	d.Line(p1, p2, color)
	d.Line(p2, p3, color)
	d.Line(p3, p1, color)
}

type debugLine struct {
	Start sprec.Vec3
	End   sprec.Vec3
	Color sprec.Vec3
}
