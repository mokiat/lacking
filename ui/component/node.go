package component

import (
	"log/slog"

	"github.com/mokiat/lacking/ui"
)

type componentNode struct {
	name        string
	element     *ui.Element
	parentScope Scope
	instance    Instance

	ref Renderable

	innerNode *componentNode
	children  []*componentNode

	isDirty bool
}

func createComponentNode(element *ui.Element, parentScope Scope, instance Instance) *componentNode {
	component := instance.component
	logger.Debug("Creating node",
		slog.String("id", instance.id),
		slog.String("name", instance.name),
		slog.String("type", component.TypeName()),
	)

	element.SetID(instance.id)

	node := &componentNode{
		name:        instance.name,
		element:     element,
		parentScope: parentScope,
		instance:    instance,
	}
	node.ref = component.Allocate(element, node.invalidate)

	scope := instance.applyScopeModifier(parentScope)
	component.HandleCreate(node.ref, scopeWithComponentNode(scope, node), instance.Properties())

	// Check if there is a inner component to delegate to.
	innerInstance := node.ref.Render()
	if innerInstance.component == nil {
		// We are currently at the leaf Element component. We need to create
		// any children now and stop the recursion.
		node.children = make([]*componentNode, len(instance.properties.children))
		for i, childInstance := range instance.properties.children {
			childElement := element.Window().CreateElement()
			child := createComponentNode(childElement, scope, childInstance)
			node.element.AppendChild(child.element)
			node.children[i] = child
		}
	} else {
		// We are still traversing the component chain so we need to go deeper.
		node.innerNode = createComponentNode(element, scope, innerInstance)
	}
	return node
}

func (node *componentNode) reconcile(instance Instance) {
	logger.Debug("Updating node",
		slog.String("name", node.name),
		slog.String("type", node.instance.component.TypeName()),
	)

	node.element.SetID(instance.id)

	component := instance.component
	if component != node.instance.component {
		panic("dynamic component chain: component type mismatch")
	}
	node.instance = instance

	// Notify that the component has been updated.
	scope := instance.applyScopeModifier(node.parentScope)
	component.HandleUpdate(node.ref, scopeWithComponentNode(scope, node), instance.Properties())

	// Check if there is a inner component to delegate to.
	if node.innerNode == nil {
		// We are currently at the leaf Element component. We need to reconcile
		// the children.
		if node.hasMatchingChildren(instance) {
			for i, childNode := range node.children {
				childNode.reconcile(instance.properties.children[i])
			}
		} else {
			for _, childNode := range node.children {
				if instance.hasMatchingChild(childNode.instance) {
					childNode.element.Detach()
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
					newChildElement := node.element.Window().CreateElement()
					newChildren[i] = createComponentNode(newChildElement, scope, childInstance)
				}
				node.element.AppendChild(newChildren[i].element)
			}
			node.children = newChildren
		}
	} else {
		// We are still traversing the component chain so we need to go deeper.
		innerInstance := node.ref.Render()
		node.innerNode.reconcile(innerInstance)
	}
}

func (node *componentNode) destroy() {
	logger.Debug("Destroying node",
		slog.String("name", node.name),
		slog.String("type", node.instance.component.TypeName()),
	)

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

	// Destroy the current component.
	component := node.instance.component
	component.HandleDelete(node.ref)
	component.Release(node.ref)

	node.element.Destroy()
	node.element = nil
}

func (node *componentNode) invalidate() {
	// TODO: Consider optimizing by tracking dirty nodes and then running
	// reconciliation only for top-level dirty nodes. Furthermore, schedule
	// only a single reconciliation call for all of them.
	if node.isDirty {
		return // already marked as dirty
	}
	node.isDirty = true
	context := node.parentScope.Context()
	context.Schedule(func() {
		if node.isValid() && node.isDirty {
			node.reconcile(node.instance)
			node.isDirty = false
		}
	})
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
	return node.element != nil
}

func scopeWithComponentNode(scope Scope, node *componentNode) Scope {
	return TypedValueScope(scope, node)
}

func componentNodeFromScope(scope Scope) *componentNode {
	return TypedValue[*componentNode](scope)
}
