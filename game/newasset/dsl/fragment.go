package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/model"
)

func CreateFragment(name string, operations ...Operation) Provider[*model.Fragment] {
	provider := OnceProvider(&fragmentProvider{
		name:       name,
		operations: operations,
	})
	fragments[name] = provider
	return provider
}

type fragmentProvider struct {
	name       string
	operations []Operation
}

func (p *fragmentProvider) Get() (*model.Fragment, error) {
	fragment := &model.Fragment{}
	fragment.SetName(p.name)
	for _, operation := range p.operations {
		if err := operation.Apply(fragment); err != nil {
			return nil, fmt.Errorf("error applying operation on %q fragment: %w", p.name, err)
		}
	}
	return fragment, nil
}

func (p *fragmentProvider) Digest() ([]byte, error) {
	return digestItems("fragment", p.name, p.operations)
}

func AddNode(nodeProvider Provider[*model.Node]) Operation {
	return &addNodeOperation{
		nodeProvider: nodeProvider,
	}
}

type addNodeOperation struct {
	nodeProvider Provider[*model.Node]
}

func (o *addNodeOperation) Apply(target any) error {
	container, ok := target.(model.NodeContainer)
	if !ok {
		return fmt.Errorf("target %T is not a node container", target)
	}
	node, err := o.nodeProvider.Get()
	if err != nil {
		return fmt.Errorf("error getting node: %w", err)
	}
	container.AddNode(node)
	return nil
}

func (o *addNodeOperation) Digest() ([]byte, error) {
	return digestItems("add-node", o.nodeProvider)
}
