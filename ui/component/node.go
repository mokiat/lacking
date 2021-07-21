package component

import "github.com/mokiat/lacking/ui"

type componentNode struct {
	instance Instance
	children []*componentNode
	element  *ui.Element

	states [][]State
}

func createComponentNode(instance Instance) *componentNode {
	result := &componentNode{
		instance: instance,
		states:   make([][]State, 1),
	}

	renderCtx = renderContext{
		node:        result,
		firstRender: true,
		lastRender:  false,
	}
	for instance.element == nil {
		instance = instance.componentFunc(instance.properties())
		renderCtx.stateDepth++
		renderCtx.stateIndex = 0
		result.states = append(result.states, nil)
	}

	result.element = instance.element
	result.children = make([]*componentNode, len(instance.children))
	for i, childInstance := range instance.children {
		child := createComponentNode(childInstance)
		instance.element.AppendChild(child.element)
		result.children[i] = child
	}

	return result
}

func (node *componentNode) destroy() {
	for _, child := range node.children {
		child.destroy()
	}
	node.children = nil

	instance := node.instance
	renderCtx = renderContext{
		node:        node,
		firstRender: false,
		lastRender:  true,
	}
	for instance.element == nil {
		instance = instance.componentFunc(instance.properties())
		renderCtx.stateDepth++
		renderCtx.stateIndex = 0
	}
	node.element = nil
}

func (node *componentNode) reconcile(instance Instance) {
	node.instance = instance
	renderCtx = renderContext{
		node:        node,
		firstRender: false,
		lastRender:  false,
	}
	for instance.element == nil {
		instance = instance.componentFunc(instance.properties())
		renderCtx.stateDepth++
		renderCtx.stateIndex = 0
	}
	if instance.element != node.element {
		panic("component chain should not return a different element instance")
	}

	if node.hasMatchingChildren(instance) {
		for i, child := range node.children {
			child.reconcile(instance.children[i])
		}
	} else {
		for _, child := range node.children {
			if instance.hasMatchingChild(child.instance) {
				child.element.Detach()
			} else {
				child.destroy()
			}
			newChildren := make([]*componentNode, len(instance.children))
			for i, childInstance := range instance.children {
				if existingChild, index := node.findChild(childInstance); index >= 0 {
					existingChild.reconcile(childInstance)
					newChildren[i] = existingChild
				} else {
					newChildren[i] = createComponentNode(childInstance)
				}
				node.element.AppendChild(newChildren[i].element)
			}
			node.children = newChildren
		}
	}
}

func (node *componentNode) hasMatchingKey(instance Instance) bool {
	return node.instance.key == instance.key
}

func (node *componentNode) hasMatchingType(instance Instance) bool {
	return node.instance.componentType == instance.componentType
}

func (node *componentNode) hasMatchingChildren(instance Instance) bool {
	if len(node.children) != len(instance.children) {
		return false
	}
	for i, child := range node.children {
		if !child.hasMatchingKey(instance.children[i]) {
			return false
		}
		if !child.hasMatchingType(instance.children[i]) {
			return false
		}
	}
	return true
}

func (node *componentNode) findChild(instance Instance) (*componentNode, int) {
	for i, child := range node.children {
		if child.hasMatchingKey(instance) && child.hasMatchingType(instance) {
			return child, i
		}
	}
	return nil, -1
}

func (node *componentNode) isValid() bool {
	return node.element != nil
}
