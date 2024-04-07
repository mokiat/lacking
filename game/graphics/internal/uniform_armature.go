package internal

import "github.com/mokiat/lacking/util/blob"

type ArmatureUniform struct {
	BoneMatrices []byte
}

func (u ArmatureUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotBytes(u.BoneMatrices)
}

func (u ArmatureUniform) Std140Size() uint32 {
	return 64 * 256
}
