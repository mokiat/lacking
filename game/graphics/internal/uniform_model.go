package internal

import "github.com/mokiat/lacking/util/blob"

type ModelUniform struct {
	ModelMatrices []byte
}

func (u ModelUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotBytes(u.ModelMatrices)
}

func (u ModelUniform) Std140Size() uint32 {
	return 64 * 256
}
