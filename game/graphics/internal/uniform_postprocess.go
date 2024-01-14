package internal

import "github.com/mokiat/lacking/util/blob"

type PostprocessUniform struct {
	Exposure float32
}

func (u PostprocessUniform) Plot(plotter *blob.Plotter, padding int) {
	// vec4
	plotter.PlotFloat32(u.Exposure)
	plotter.Skip(4 + 4 + 4)

	plotter.Skip(padding)
}

func (u PostprocessUniform) Std140Size() int {
	return 16
}
