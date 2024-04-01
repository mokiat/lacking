package dsl

import "github.com/mokiat/lacking/game/newasset/model"

func CreateNode(name string, operations ...Operation) Provider[*model.BaseNode] {
	get := func() (*model.BaseNode, error) {
		node := &model.BaseNode{}
		node.SetName(name)
		for _, op := range operations {
			if err := op.Apply(node); err != nil {
				return nil, err
			}
		}
		return node, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("node", name, operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}
