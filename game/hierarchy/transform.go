package hierarchy

import "github.com/mokiat/gomath/dprec"

// TransformFunc is a mechanism to calculate a custom absolute matrix
// for the node.
type TransformFunc func(node *Node) dprec.Mat4

// DefaultTransformFunc is a TransformFunc that applies standard matrix
// multiplication rules.
func DefaultTransformFunc(node *Node) dprec.Mat4 {
	return dprec.Mat4Prod(node.BaseAbsoluteMatrix(), node.Matrix())
}
