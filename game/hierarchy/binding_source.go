package hierarchy

// SourceBinding represents a binding between a source of data to a node.
type SourceBinding[T BindingObject] interface {
	Binding[T]

	// OnSourceToNode is called when the source's state should be applied to the
	// node.
	OnSourceToNode(*Scene, T, NodeID)
}

// NewSourceBindingSet creates a binding set that tracks deletions and notifies
// when the source should be applied to the node.
func NewSourceBindingSet[T BindingObject](scene *Scene, binding SourceBinding[T]) *SourceBindingSet[T] {
	result := &SourceBindingSet[T]{
		BindingSet:    NewBindingSet(scene, binding),
		sourceBinding: binding,
	}
	scene.SubscribeSourceApply(func(s *Scene, id NodeID) {
		result.ApplySourceToNode(id)
	})
	return result
}

// SourceBindingSet represents a set of source bindings for a specific object
// type.
type SourceBindingSet[T BindingObject] struct {
	*BindingSet[T]
	sourceBinding SourceBinding[T]
}

// ApplyTargetToNode applies the state of the target to its node.
func (s *SourceBindingSet[T]) ApplySourceToNode(id NodeID) {
	if target, ok := s.relations[id]; ok {
		s.sourceBinding.OnSourceToNode(s.scene, target, id)
	}
}
