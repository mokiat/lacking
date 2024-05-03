package dsl

import (
	"github.com/mokiat/lacking/game/asset/mdl"
)

// CreateNode creates a new node with the specified name and operations.
func CreateNode(name string, operations ...Operation) Provider[*mdl.Node] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Node, error) {
			node := mdl.NewNode(name)
			for _, op := range operations {
				if err := op.Apply(node); err != nil {
					return nil, err
				}
			}
			return node, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-node", name, operations)
		},
	))
}
