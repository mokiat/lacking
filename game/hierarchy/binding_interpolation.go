package hierarchy

// InterpolationBinding represents a binding between a node and a target that
// needs to use interpolated state from the node.
type InterpolationBinding[T BindingObject] interface {
	Binding[T]

	// OnNodeToInterpolation is called when the node's state should be applied to
	// the target, using interpolation.
	OnNodeToInterpolation(*Scene, NodeID, T, float64)
}

// InterpolationBindingSet creates a binding set that tracks deletions and
// notifies when the node should be applied to the target using interpolation.
func NewInterpolationBindingSet[T BindingObject](scene *Scene, binding InterpolationBinding[T]) *InterpolationBindingSet[T] {
	result := &InterpolationBindingSet[T]{
		BindingSet:           NewBindingSet(scene, binding),
		interpolationBinding: binding,
	}
	scene.SubscribeInterpolationApply(func(s *Scene, id NodeID, fraction float64) {
		result.ApplyNodeToInterpolation(id, fraction)
	})
	return result
}

// InterpolationBindingSet represents a set of interpolation bindings for a
// specific object type.
type InterpolationBindingSet[T BindingObject] struct {
	*BindingSet[T]
	interpolationBinding InterpolationBinding[T]
}

// ApplyNodeToInterpolation applies the state of the node to its target using
// interpolation.
func (s *InterpolationBindingSet[T]) ApplyNodeToInterpolation(id NodeID, fraction float64) {
	if target, ok := s.relations[id]; ok {
		s.interpolationBinding.OnNodeToInterpolation(s.scene, id, target, fraction)
	}
}
