package hierarchy

// TargetBinding represents a binding between a node and a target.
type TargetBinding[T BindingObject] interface {
	Binding[T]

	// OnNodeToTarget is called when the node's state should be applied to the
	// target.
	OnNodeToTarget(*Scene, NodeID, T)
}

// NewTargetBindingSet creates a binding set that tracks deletions and notifies
// when the node should be applied to the target.
func NewTargetBindingSet[T BindingObject](scene *Scene, binding TargetBinding[T]) *TargetBindingSet[T] {
	result := &TargetBindingSet[T]{
		BindingSet:    NewBindingSet(scene, binding),
		targetBinding: binding,
	}
	scene.SubscribeTargetApply(func(s *Scene, id NodeID) {
		result.ApplyNodeToTarget(id)
	})
	return result
}

// TargetBindingSet represents a set of target bindings for a specific object
// type.
type TargetBindingSet[T BindingObject] struct {
	*BindingSet[T]
	targetBinding TargetBinding[T]
}

// ApplyNodeToTarget applies the state of the node to its target.
func (s *TargetBindingSet[T]) ApplyNodeToTarget(id NodeID) {
	if target, ok := s.relations[id]; ok {
		s.targetBinding.OnNodeToTarget(s.scene, id, target)
	}
}
