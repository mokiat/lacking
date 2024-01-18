package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type SkyboxUniform struct {
	Color sprec.Vec4
}

func (u SkyboxUniform) Std140Plot(plotter *blob.Plotter) {
	// vec4
	plotter.PlotSPVec4(u.Color)
}

func (u SkyboxUniform) Std140Size() int {
	return 16
}
