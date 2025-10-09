package hierarchy

const (
	initialBindingCapacity = 64
)

// Binding represents a relationship between a target and a node in the scene.
type Binding[T BindingTarget] interface {

	// OnTargetToNode is called when the target's state should be applied to the
	// node.
	OnTargetToNode(*Scene, T, NodeID)

	// OnNodeToTarget is called when the node's state should be applied to the
	// target.
	OnNodeToTarget(*Scene, NodeID, T, float64)

	// OnStaleBinding is called when the node is deleted and the binding
	// is determined to be no longer valid.
	OnStaleBinding(*Scene, T)
}

// BindingTarget represents a type that can be used as a target in a binding.
type BindingTarget interface {
	comparable
}

// NewBindingSet represents a set of bindings for a specific target type.
func NewBindingSet[T BindingTarget](scene *Scene, binding Binding[T]) *BindingSet[T] {
	result := &BindingSet[T]{
		scene:     scene,
		binding:   binding,
		relations: make(map[NodeID]T, initialBindingCapacity),
	}
	scene.SubscribeNodeDelete(func(s *Scene, id NodeID) {
		if target, ok := result.Unbind(id); ok {
			binding.OnStaleBinding(s, target)
		}
	})
	return result
}

// BindingSet represents a set of bindings for a specific target type.
type BindingSet[T BindingTarget] struct {
	scene     *Scene
	binding   Binding[T]
	relations map[NodeID]T
}

// Bind binds the target to the node with the given ID.
func (s *BindingSet[T]) Bind(id NodeID, target T) {
	if s.scene.IsValidNode(id) {
		s.relations[id] = target
	}
}

// Unbind unbinds the target from its node.
func (s *BindingSet[T]) Unbind(id NodeID) (T, bool) {
	target, exists := s.relations[id]
	if exists {
		delete(s.relations, id)
	}
	return target, exists
}

// ApplyTargetToNode applies the state of the target to its node.
func (s *BindingSet[T]) ApplyTargetToNode(id NodeID) {
	if target, ok := s.relations[id]; ok {
		s.binding.OnTargetToNode(s.scene, target, id)
	}
}

// ApplyTargetsToNodes applies the state of the targets to their nodes.
func (s *BindingSet[T]) ApplyTargetsToNodes() {
	for id, target := range s.relations {
		if s.scene.IsValidNode(id) {
			s.binding.OnTargetToNode(s.scene, target, id)
		}
	}
}

// ApplyNodeToTarget applies the state of the node to its target.
func (s *BindingSet[T]) ApplyNodeToTarget(id NodeID, fraction float64) {
	if target, ok := s.relations[id]; ok {
		s.binding.OnNodeToTarget(s.scene, id, target, fraction)
	}
}

// ApplyNodesToTargets applies the state of the nodes to their targets.
func (s *BindingSet[T]) ApplyNodesToTargets(fraction float64) {
	for id, target := range s.relations {
		if s.scene.IsValidNode(id) {
			s.binding.OnNodeToTarget(s.scene, id, target, fraction)
		}
	}
}

// DeleteStale removes bindings to nodes that are no longer valid.
func (s *BindingSet[T]) DeleteStale() {
	for id, target := range s.relations {
		if !s.scene.IsValidNode(id) {
			delete(s.relations, id)
			s.binding.OnStaleBinding(s.scene, target)
		}
	}
}
