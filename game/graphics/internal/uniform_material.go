package internal

import "github.com/mokiat/lacking/util/blob"

type MaterialUniform struct {
	Data []byte
}

func (u MaterialUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotBytes(u.Data)
}

func (u MaterialUniform) Std140Size() int {
	return len(u.Data)
}
