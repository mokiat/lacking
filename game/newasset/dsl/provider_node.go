package dsl

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

// CreateNode creates a new node with the specified name and operations.
func CreateNode(name string, operations ...Operation) Provider[*mdl.BaseNode] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.BaseNode, error) {
			var node mdl.BaseNode
			node.SetName(name)
			node.SetTranslation(dprec.ZeroVec3())
			node.SetRotation(dprec.IdentityQuat())
			node.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
			for _, op := range operations {
				if err := op.Apply(&node); err != nil {
					return nil, err
				}
			}
			return &node, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-node", name, operations)
		},
	))
}
