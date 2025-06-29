package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type TimingUniform struct {
	Vectors [256]sprec.Vec4
}

func (u TimingUniform) Std140Plot(plotter *blob.Plotter) {
	for _, vector := range u.Vectors {
		plotter.PlotSPVec4(vector)
	}
}

func (u TimingUniform) Std140Size() uint32 {
	return 256 * 16
}
