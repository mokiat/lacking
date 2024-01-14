package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type SkyboxUniform struct {
	Color sprec.Vec4
}

func (u SkyboxUniform) Plot(plotter *blob.Plotter, padding int) {
	// vec4
	plotter.PlotSPVec4(u.Color)

	plotter.Skip(padding)
}

func (u SkyboxUniform) Std140Size() int {
	return 16
}
