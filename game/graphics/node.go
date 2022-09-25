package graphics

import (
	"encoding/binary"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/util/blob"
)

func newNode() *Node {
	matrixData := make([]byte, 64)
	plotter := blob.NewPlotter(matrixData)
	plotter.PlotSPMat4(sprec.IdentityMat4())
	return &Node{
		matrixData: matrixData,
	}
}

// Node represents a positioning of some entity in
// the 3D scene.
type Node struct {
	matrixData []byte
}

// SetMatrix changes the model matrix of this node.
// Keep in mind that this is a somewhat slow operation and should only
// be performed only once per frame. This is also the reason why there is
// no getter for this method. Clients are expected to track matrices outside
// this type if needed.
func (n *Node) SetMatrix(matrix dprec.Mat4) {
	plotter := blob.NewPlotter(n.matrixData)
	// plotter.PlotMat4(dtos.Mat4(matrix))
	plotter.PlotFloat32(float32(matrix.M11))
	plotter.PlotFloat32(float32(matrix.M21))
	plotter.PlotFloat32(float32(matrix.M31))
	plotter.Skip(4) // skip matrix.M41 (assume unchanged)
	plotter.PlotFloat32(float32(matrix.M12))
	plotter.PlotFloat32(float32(matrix.M22))
	plotter.PlotFloat32(float32(matrix.M32))
	plotter.Skip(4) // skip matrix.M42 (assume unchanged)
	plotter.PlotFloat32(float32(matrix.M13))
	plotter.PlotFloat32(float32(matrix.M23))
	plotter.PlotFloat32(float32(matrix.M33))
	plotter.Skip(4) // skip matrix.M43 (assume unchanged)
	plotter.PlotFloat32(float32(matrix.M14))
	plotter.PlotFloat32(float32(matrix.M24))
	plotter.PlotFloat32(float32(matrix.M34))
	plotter.Skip(4) // skip matrix.M44 (assume unchanged)
}

func (n *Node) gfxMatrix() sprec.Mat4 {
	scanner := buffer.NewScanner(n.matrixData, binary.LittleEndian)
	return scanner.ScanMat4()
}
