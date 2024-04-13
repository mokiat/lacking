package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func AddNode(nodeProvider Provider[mdl.Node]) Operation {
	apply := func(target any) error {
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
	}

	digest := func() ([]byte, error) {
		return CreateDigest("add-node", nodeProvider)
	}

	return FuncOperation(apply, digest)
}
