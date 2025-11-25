package hierarchy

import "github.com/mokiat/gomath/dprec"

// TransformFunc is a mechanism to calculate a custom absolute matrix
// for the node from its local state.
type TransformFunc func(scene *Scene, node NodeID) dprec.Mat4

// DefaultTransformFunc is a TransformFunc that applies standard matrix
// multiplication rules.
func DefaultTransformFunc(scene *Scene, node NodeID) dprec.Mat4 {
	baseMatrix := scene.NodeBaseMatrix(node)
	nodeMatrix := scene.NodeMatrix(node)
	return dprec.Mat4Prod(baseMatrix, nodeMatrix)
}
