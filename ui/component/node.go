package component

import "github.com/mokiat/lacking/ui"

type componentNodeKey struct{}

type componentNode struct {
	outerScope Scope
	instance   Instance
	ref        Renderable

	innerNode *componentNode
	children  []*componentNode
	element   *ui.Element

	isDirty bool
}

func createComponentNode(scope Scope, instance Instance) *componentNode {
	component := instance.component
	logger.Debug("Creating node (type: %q).", component.TypeName())

	if instance.scope != nil {
		scope = instance.scope
	}
	node := &componentNode{
		outerScope: scope,
		instance:   instance,
	}
	node.ref = component.Allocate(node, node.invalidate)

	// Notify that the component has been created.
	component.NotifyCreate(node.ref, instance.Properties())

	// Check if there is a inner component to delegate to.
	// In reality this check fails only once an Element component is reached.
	innerInstance := node.ref.Render()
	if innerInstance.element == nil {
		node.innerNode = createComponentNode(node, innerInstance)
		return node // nothing more to do
	}
	node.element = innerInstance.element

	// We are the leaf Element component, so we need to do element and children
	// management.
	node.children = make([]*componentNode, len(instance.properties.children))
	for i, childInstance := range instance.properties.children {
		child := createComponentNode(node, childInstance)
		node.leafElement().AppendChild(child.leafElement())
		node.children[i] = child
	}
	return node
}

func (node *componentNode) reconcile(instance Instance) {
	logger.Debug("Updating node (type: %q).", node.instance.component.TypeName())

	component := instance.component
	if component != node.instance.component {
		panic("dynamic component chain: component type mismatch")
	}
	node.instance = instance

	// Notify that the component has been updated.
	component.NotifyUpdate(node.ref, instance.Properties())

	// Check if there is a inner component to delegate to.
	innerInstance := node.ref.Render()
	if node.innerNode != nil {
		node.innerNode.reconcile(innerInstance)
		return // nothing more to do
	}
	if innerInstance.element != node.element {
		panic("dynamic component chain: element mismatch")
	}

	// We are the leaf Element component, so we need to do element and children
	// management.
	if node.hasMatchingChildren(instance) {
		for i, childNode := range node.children {
			childNode.reconcile(instance.properties.children[i])
		}
	} else {
		for _, childNode := range node.children {
			if instance.hasMatchingChild(childNode.instance) {
				childNode.leafElement().Detach()
			} else {
				childNode.destroy()
			}
		}
		newChildren := make([]*componentNode, len(instance.properties.children))
		for i, childInstance := range instance.properties.children {
			if existingChildNode, index := node.findChild(childInstance); index >= 0 {
				existingChildNode.reconcile(childInstance)
				newChildren[i] = existingChildNode
			} else {
				newChildren[i] = createComponentNode(node, childInstance)
			}
			node.leafElement().AppendChild(newChildren[i].leafElement())
		}
		node.children = newChildren
	}
}

func (node *componentNode) destroy() {
	logger.Debug("Destroying node (type: %q).", node.instance.component.TypeName())

	// Start by destroying nested components first.
	if node.innerNode != nil {
		node.innerNode.destroy()
		node.innerNode = nil
	}

	// Destroy any children, if there are such. This will only loop for
	// leaf Element component nodes.
	for _, childNode := range node.children {
		childNode.destroy()
	}
	node.children = nil

	// Notify that the component has been destroyed. The Element component
	// implementation will automatically detach and destroy the ui Element.
	component := node.instance.component
	component.NotifyDelete(node.ref)

	node.element = nil
	component.Release(node.ref)
}

func (node *componentNode) leafElement() *ui.Element {
	if node.innerNode != nil {
		return node.innerNode.leafElement()
	}
	return node.element
}

func (node *componentNode) invalidate() {
	// TODO: Optimize by grouping and sorting so that lower depth components
	// are invalidated first, making higher level component reconciliations
	// no-op in most cases.
	if !node.isDirty {
		node.isDirty = true
		node.Context().Schedule(func() {
			if node.isValid() && node.isDirty {
				node.reconcile(node.instance)
				node.isDirty = false
			}
		})
	}
}

func (node *componentNode) hasMatchingKey(instance Instance) bool {
	return node.instance.key == instance.key
}

func (node *componentNode) hasMatchingType(instance Instance) bool {
	return node.instance.component == instance.component
}

func (node *componentNode) hasMatchingChildren(instance Instance) bool {
	if len(node.children) != len(instance.properties.children) {
		return false
	}
	for i, child := range node.children {
		if !child.hasMatchingKey(instance.properties.children[i]) {
			return false
		}
		if !child.hasMatchingType(instance.properties.children[i]) {
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
	return node.leafElement() != nil
}

func (node *componentNode) Context() *ui.Context {
	return node.outerScope.Context()
}

func (node *componentNode) Value(key any) any {
	if key == (componentNodeKey{}) {
		return node
	}
	return node.outerScope.Value(key)
}
