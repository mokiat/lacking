package component

import (
	"github.com/mokiat/lacking/ui"
)

type dirtiable interface {
	isDirty() bool
	setDirty(bool)
}

type componentNode struct {
	instance Instance
	children []*componentNode
	element  *ui.Element
	scope    Scope

	states [][]dirtiable
}

func createComponentNode(instance Instance, scope Scope) *componentNode {
	result := &componentNode{
		instance: instance,
		scope:    scope,
		states:   make([][]dirtiable, 1),
	}

	renderCtx = renderContext{
		node:        result,
		firstRender: true,
		lastRender:  false,
		properties:  instance.properties(),
	}
	for instance.element == nil {
		if instance.scope != nil {
			result.scope = instance.scope
		}
		instance = instance.componentFunc(instance.properties(), result.scope)
		renderCtx.stateDepth++
		renderCtx.stateIndex = 0
		renderCtx.properties = instance.properties()
		result.states = append(result.states, nil)
	}

	result.element = instance.element
	result.children = make([]*componentNode, len(instance.children))
	for i, childInstance := range instance.children {
		child := createComponentNode(childInstance, result.scope)
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
		instance = instance.componentFunc(instance.properties(), node.scope)
		renderCtx.stateDepth++
		renderCtx.stateIndex = 0
	}
	node.element = nil
}

func (node *componentNode) reconcile(instance Instance, scope Scope) {
	node.instance = instance
	node.scope = scope
	renderCtx = renderContext{
		node:         node,
		firstRender:  false,
		lastRender:   false,
		forcedRender: node.consumeDirty(),
		properties:   instance.properties(),
	}
	for instance.element == nil {
		if instance.scope != nil {
			node.scope = instance.scope
		}
		instance = instance.componentFunc(instance.properties(), node.scope)
		renderCtx.stateDepth++
		renderCtx.stateIndex = 0
		renderCtx.properties = instance.properties()
	}
	if instance.element != node.element {
		panic("component chain should not return a different element instance")
	}

	if node.hasMatchingChildren(instance) {
		for i, child := range node.children {
			child.reconcile(instance.children[i], node.scope)
		}
	} else {
		for _, child := range node.children {
			if instance.hasMatchingChild(child.instance) {
				child.element.Detach()
			} else {
				child.destroy()
			}
		}
		newChildren := make([]*componentNode, len(instance.children))
		for i, childInstance := range instance.children {
			if existingChild, index := node.findChild(childInstance); index >= 0 {
				existingChild.reconcile(childInstance, node.scope)
				newChildren[i] = existingChild
			} else {
				newChildren[i] = createComponentNode(childInstance, node.scope)
			}
			node.element.AppendChild(newChildren[i].element)
		}
		node.children = newChildren
	}
}

func (node *componentNode) consumeDirty() bool {
	var dirty = false
	for _, depth := range node.states {
		for _, state := range depth {
			if state.isDirty() {
				dirty = true
				state.setDirty(false)
			}
		}
	}
	return dirty
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
