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

func SetContent[T any](provider Provider[T]) Operation {
	apply := func(target any) error {
		contentable, ok := target.(model.ContentHolder)
		if !ok {
			return fmt.Errorf("target %T is not a content holder", target)
		}
		content, err := provider.Get()
		if err != nil {
			return fmt.Errorf("error getting content: %w", err)
		}
		contentable.SetContent(content)
		return nil
	}

	digest := func() ([]byte, error) {
		return digestItems("set-content", provider)
	}

	return FuncOperation(apply, digest)
}
