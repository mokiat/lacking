package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// SetSource sets the transformation source of the node.
func SetSource[T any](nodeSourceProvider Provider[T]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			nodeSource, err := nodeSourceProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting node source: %w", err)
			}
			node, ok := target.(*mdl.Node)
			if !ok {
				return fmt.Errorf("target %T is not a node", target)
			}
			node.SetSource(nodeSource)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-source", nodeSourceProvider)
		},
	)
}

// SetTarget sets the transformation target of the node.
func SetTarget[T any](nodeTargetProvider Provider[T]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			nodeTarget, err := nodeTargetProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting node target: %w", err)
			}
			node, ok := target.(*mdl.Node)
			if !ok {
				return fmt.Errorf("target %T is not a node", target)
			}
			node.SetTarget(nodeTarget)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-target", nodeTargetProvider)
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
