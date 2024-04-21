package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// AddNode adds the specified node to the target node container.
func AddNode(nodeProvider Provider[mdl.Node]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			container, ok := target.(mdl.NodeContainer)
			if !ok {
				return fmt.Errorf("target %T is not a node container", target)
			}

			node, err := nodeProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting node: %w", err)
			}
			container.AddNode(node)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("add-node", nodeProvider)
		},
	)
}
