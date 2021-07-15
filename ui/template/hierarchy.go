package template

import (
	"github.com/mokiat/lacking/ui"
)

type hierarchy struct{}

func (s *hierarchy) CreateComponentNode(instance Instance) *componentNode {
	var result *componentNode

	if instance.element == nil {
		result = &componentNode{
			key:           instance.key,
			componentType: instance.componentType,
			componentFunc: instance.componentFunc,
			instance:      instance,
		}

		renderCtx = renderContext{
			node:        result,
			firstRender: true,
			lastRender:  false,
		}
		nextInstance := instance.componentFunc(instance.properties())
		result.next = s.CreateComponentNode(nextInstance)
	} else {
		result = &componentNode{
			element:  instance.element,
			children: make([]*componentNode, len(instance.children)),
		}
		for i, childInstance := range instance.children {
			child := s.CreateComponentNode(childInstance)
			instance.element.AppendChild(child.Element())
			result.children[i] = child
		}
	}

	return result
}

func (s *hierarchy) DestroyComponentNode(node *componentNode) {
	if node.next != nil {
		s.DestroyComponentNode(node.next)
	}
	for _, child := range node.children {
		s.DestroyComponentNode(child)
	}
	node.children = nil

	if node.componentFunc != nil {
		renderCtx = renderContext{
			firstRender: false,
			lastRender:  true,
		}
		node.componentFunc(node.instance.properties())
		node.componentFunc = nil
	}
}

func (s *hierarchy) Reconcile(node *componentNode, instance Instance) {
	// 	node.reconcilable = reconcilable
	// 	switch {
	// 	case reconcilable.isBranch():
	// 		s.ReconcileBranch(node, reconcilable.branch)
	// 	case reconcilable.isLeaf():
	// 		s.ReconcileLeaf(node, reconcilable.leaf)
	// 	}
}

// func (s *hierarchy) ReconcileBranch(node *componentNode, branch *reconcilableBranch) {
// 	log.Println("reconcile branch")

// 	if !node.IsDirty() {
// 		return
// 	}

// 	if node.next == nil {
// 		panic("NOT BRANCH: TODO: Should recreate this one fully")
// 	}

// 	var needsRender = false
// 	if !node.HasMatchingData(branch.instance) {
// 		needsRender = true
// 		node.component.OnDataChanged(branch.instance.data)
// 	}
// 	if !node.HasMatchingLayoutData(branch.instance) {
// 		needsRender = true
// 		node.component.OnLayoutDataChanged(branch.instance.layoutData)
// 	}
// 	if !node.HasMatchingChildren(branch.instance) {
// 		needsRender = true
// 		node.component.OnChildrenChanged(branch.instance.children)
// 	}
// 	if needsRender {
// 		nextReconcilable := node.component.Render(RenderContext{
// 			Context:      s.ctx,
// 			Key:          branch.instance.key,
// 			Data:         branch.instance.data,
// 			LayoutData:   branch.instance.layoutData,
// 			CallbackData: branch.instance.callbackData,
// 			Children:     branch.instance.children,
// 		})
// 		s.Reconcile(node.next, nextReconcilable)
// 	} else {
// 		for _, child := range node.children {
// 			s.Reconcile(child, child.reconcilable)
// 		}
// 	}
// }

// func (s *hierarchy) ReconcileLeaf(node *componentNode, leaf *reconcilableLeaf) {
// 	log.Println("reconcile leaf")

// 	if !node.IsDirty() {
// 		return
// 	}

// 	if node.next != nil {
// 		panic("NOT LEAF: TODO: Should recreate this one fully")
// 	}

// 	// detach current children
// 	for _, child := range node.children {
// 		child.element.Detach()
// 	}

// 	newChildren := make([]*componentNode, len(leaf.children))
// 	// reuse or create new children
// 	for i, childInstance := range leaf.children {
// 		if existingChildNode, index := node.FindChild(childInstance); index >= 0 {
// 			node.Element().AppendChild(existingChildNode.Element())
// 			s.Reconcile(existingChildNode, existingChildNode.reconcilable)
// 			node.children[index] = nil // ensure we don't destroy the component
// 			newChildren[i] = existingChildNode
// 		} else {
// 			childNode := s.CreateComponentNode(childInstance)
// 			node.Element().AppendChild(childNode.Element())
// 			newChildren[i] = childNode
// 		}
// 	}
// 	// destroy unmatched old children
// 	for _, child := range node.children {
// 		if child != nil {
// 			s.DestroyComponentNode(child)
// 		}
// 	}
// 	node.children = newChildren
// }

type componentNode struct {
	next *componentNode

	key           string
	componentType string
	componentFunc ComponentFunc
	instance      Instance

	states []State

	// children contains the immediate (flattened) children of a component
	// and is only applicable to the last component in a component chain
	children []*componentNode

	element *ui.Element
}

func (n *componentNode) Element() *ui.Element {
	current := n
	for current.element == nil {
		current = current.next
	}
	return current.element
}

func (n *componentNode) HasMatchingKey(instance Instance) bool {
	return n.key == instance.key
}

func (n *componentNode) HasMatchingType(instance Instance) bool {
	return n.componentType == instance.componentType
}

func (n *componentNode) HasMatchingChildren(instance Instance) bool {
	if len(n.children) != len(instance.children) {
		return false
	}
	for i := range n.children {
		if !n.HasMatchingKey(instance.children[i]) {
			return false
		}
		if !n.children[i].HasMatchingType(instance.children[i]) {
			return false
		}
	}
	return true
}

func (n *componentNode) FindChild(instance Instance) (*componentNode, int) {
	for i, child := range n.children {
		if child.HasMatchingKey(instance) && child.HasMatchingType(instance) {
			return child, i
		}
	}
	return nil, -1
}
