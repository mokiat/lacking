package graphics

import "github.com/mokiat/gomath/sprec"

// TODO: This node should be uniform-oritented
// meaning that it should store its transformation as a matrix
// directly in byte sequence ready to be uploaded to GPU.
// That way also the byte slice can be passed as reference to an update
// command without having to allocate temp memory.
// (This also means that the node should not be updated while rendering
// is going on)

func newNode() *Node {
	return &Node{
		matrix: sprec.IdentityMat4(),
	}
}

// Node represents a positioning of some entity in
// the 3D scene.
type Node struct {
	matrix sprec.Mat4
}

// SetMatrix changes the model matrix of this node.
// Keep in mind that this is a somewhat slower operation and should only
// be performed only once per frame. This is also the reason why there is
// no getter for this method.
// Clients are expected to track matrices outside this type if needed.
func (n *Node) SetMatrix(matrix sprec.Mat4) {
	// TODO: Write matrix to byte buffer to be used by a uniform buffer
	n.matrix = matrix
}

func (n *Node) innerMat() sprec.Mat4 {
	return n.matrix
}

// TODO: Byte slice to be used for upload instead.
func (n *Node) matrixArray() [16]float32 {
	return n.matrix.ColumnMajorArray()
}

func (n *Node) inverseMatrixArray() [16]float32 {
	return sprec.InverseMat4(n.matrix).ColumnMajorArray()
}
