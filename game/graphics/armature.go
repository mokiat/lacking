package graphics

import (
	"encoding/binary"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/data/buffer"
)

type ArmatureTemplateDefinition struct {
	BoneInverseMatrices []sprec.Mat4
}

type ArmatureTemplate struct {
	inverseMatrices []sprec.Mat4
}

func (t *ArmatureTemplate) boneCount() int {
	return len(t.inverseMatrices)
}

type Armature struct {
	template          *ArmatureTemplate
	uniformBufferData data.Buffer
}

func (a *Armature) BoneCount() int {
	return a.template.boneCount()
}

func (a *Armature) SetBone(index int, matrix sprec.Mat4) {
	finalMatrix := sprec.Mat4Prod(
		matrix,
		a.template.inverseMatrices[index],
	)
	plotter := buffer.NewPlotter(a.uniformBufferData, binary.LittleEndian)
	plotter.Seek(index * 64)
	plotter.PlotMat4(finalMatrix)
}
