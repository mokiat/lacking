package dsl

import (
	"github.com/mokiat/lacking/game/newasset/model"
)

func CreateNode(name string, operations ...Operation) Provider[*model.Node] {
	return OnceProvider(&nodeProvider{
		name:       name,
		operations: operations,
	})
}

type nodeProvider struct {
	name       string
	operations []Operation
}

func (p *nodeProvider) Get() (*model.Node, error) {
	node := &model.Node{}
	node.SetName(p.name)
	for _, op := range p.operations {
		if err := op.Apply(node); err != nil {
			return nil, err
		}
	}
	return node, nil
}

func (p *nodeProvider) Digest() ([]byte, error) {
	return digestItems("node", p.name, p.operations)
}
