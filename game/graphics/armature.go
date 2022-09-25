package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type ArmatureInfo struct {
	InverseMatrices []sprec.Mat4
}

type Armature struct {
	inverseMatrices   []sprec.Mat4
	uniformBufferData blob.Buffer
}

func (a *Armature) BoneCount() int {
	return len(a.inverseMatrices)
}

func (a *Armature) SetBone(index int, matrix sprec.Mat4) {
	finalMatrix := sprec.Mat4MultiProd(
		matrix,
		a.inverseMatrices[index],
	)
	plotter := blob.NewPlotter(a.uniformBufferData)
	plotter.Seek(index * 64)
	plotter.PlotSPMat4(finalMatrix)
}
