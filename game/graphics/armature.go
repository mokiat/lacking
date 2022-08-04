package graphics

import (
	"encoding/binary"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/data/buffer"
)

type ArmatureTemplate struct {
	InverseMatrices []sprec.Mat4
}

type Armature struct {
	template          ArmatureTemplate
	uniformBufferData data.Buffer
}

func (a *Armature) BoneCount() int {
	return len(a.template.InverseMatrices)
}

func (a *Armature) SetBone(index int, matrix sprec.Mat4) {
	finalMatrix := sprec.Mat4MultiProd(
		matrix,
		a.template.InverseMatrices[index],
	)
	plotter := buffer.NewPlotter(a.uniformBufferData, binary.LittleEndian)
	plotter.Seek(index * 64)
	plotter.PlotMat4(finalMatrix)
}
