package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/model"
)

func AddNode(nodeProvider Provider[*model.Node]) Operation {
	apply := func(target any) error {
		container, ok := target.(model.NodeContainer)
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
		return digestItems("add-node", nodeProvider)
	}

	return FuncOperation(apply, digest)
}
