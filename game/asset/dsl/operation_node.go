package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// AddAttachment adds an attachment to a node.
func AddAttachment[T any](attachmentProvider Provider[T]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			attachment, err := attachmentProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting attachment: %w", err)
			}
			node, ok := target.(*mdl.Node)
			if !ok {
				return fmt.Errorf("target %T is not a node", target)
			}
			node.AddAttachment(attachment)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("add-attachment", attachmentProvider)
		},
	)
}

// AddNode adds the specified node to the target node container.
func AddNode(nodeProvider Provider[*mdl.Node]) Operation {
	type nodeContainer interface {
		AddNode(*mdl.Node)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			container, ok := target.(nodeContainer)
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
