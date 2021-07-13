package template

import (
	"github.com/mokiat/lacking/ui"
)

type hierarchy struct {
	uiCtx *ui.Context
}

// TODO: Add component to list of dependents on the instance owner
// this might require that dependends be made a linked list

func (s *hierarchy) CreateComponentNode(instance Instance) *componentNode {
	var result *componentNode

	if instance.componentType != nil {
		component := instance.componentType.NewComponent(s.uiCtx)
		component.OnDataChanged(instance.data)
		component.OnLayoutDataChanged(instance.layoutData)
		component.OnCallbackDataChanged(instance.callbackData)
		component.OnChildrenChanged(instance.children)
		component.OnCreated()

		result = &componentNode{
			next:          nil,
			component:     component,
			key:           instance.key,
			componentType: instance.componentType,
			data:          instance.data,
			layoutData:    instance.layoutData,
		}

		nextInstance := component.Render(s.RenderContext(result, instance))
		result.next = s.CreateComponentNode(nextInstance)
	} else {
		result = &componentNode{
			element:  instance.element,
			children: make([]*componentNode, len(instance.children)),
		}
		for i, childInstance := range instance.children {
			child := s.CreateComponentNode(childInstance)
			result.element.AppendChild(child.Element())
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
	node.dependents = nil

	if node.component != nil {
		node.component.OnDestroyed()
		node.component = nil
	}
}

func (s *hierarchy) RenderContext(owner *componentNode, instance Instance) RenderContext {
	return RenderContext{
		Context: &Context{
			owner: owner,
		},
		instance: instance,
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

	element *ui.Element

	key           string
	componentType ComponentType
	component     Component

	data       interface{}
	layoutData interface{}

	// dirty indicates whether this component should be reconciled
	dirty bool

	// dependents contains the children that this component
	// created and as part of its render function and that are
	// potentially affected by data changes to this component.
	dependents []*componentNode

	// children contains the immediate (flattened) children of a component
	// and is only applicable to the last component in a component chain
	children []*componentNode
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
	return n.componentType.Name() == instance.componentType.Name()
}

func (n *componentNode) HasMatchingData(instance Instance) bool {
	return isDataEqual(n.data, instance.data)
}

func (n *componentNode) HasMatchingLayoutData(instance Instance) bool {
	return isLayoutDataEqual(n.layoutData, instance.layoutData)
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

func (n *componentNode) MarkDirty() {
	n.dirty = true
	for _, depenent := range n.dependents {
		depenent.MarkDirty()
	}
}

func (n *componentNode) MarkClean() {
	n.dirty = false
}

func (n *componentNode) IsDirty() bool {
	return n.dirty
}

func (n *componentNode) FindChild(instance Instance) (*componentNode, int) {
	for i, child := range n.children {
		if child.HasMatchingKey(instance) && child.HasMatchingType(instance) {
			return child, i
		}
	}
	return nil, -1
}
