package internal

import "github.com/mokiat/lacking/util/blob"

type PostprocessUniform struct {
	Exposure float32
}

func (u PostprocessUniform) Std140Plot(plotter *blob.Plotter) {
	// vec4
	plotter.PlotFloat32(u.Exposure)
	plotter.Skip(4 + 4 + 4)
}

func (u PostprocessUniform) Std140Size() uint32 {
	return 16
}
