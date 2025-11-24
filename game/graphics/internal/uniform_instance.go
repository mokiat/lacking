package internal

import (
	"github.com/mokiat/lacking/util/blob"
)

type InstanceUniform struct {
	InstanceBlocks []byte
}

func (u InstanceUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotBytes(u.InstanceBlocks)
}

func (u InstanceUniform) Std140Size() uint32 {
	return 256 * 16
}
