package hierarchy

import "github.com/mokiat/gog"

const (
	initialBindingCapacity = 64
)

// BindingObject represents a type that can be bound to a node.
type BindingObject interface {
	comparable
}

// Binding represents a relationship between an object and a node.
type Binding[T BindingObject] interface {

	// OnStaleBinding is called when the node is deleted and the binding
	// is determined to no longer be valid.
	OnStaleBinding(*Scene, T)
}

// NewManualBindingSet creates a binding set that only tracks deletions.
func NewBindingSet[T BindingObject](scene *Scene, binding Binding[T]) *BindingSet[T] {
	result := &BindingSet[T]{
		scene:     scene,
		binding:   binding,
		relations: make(map[NodeID]T, initialBindingCapacity),
	}
	scene.SubscribeNodeDelete(func(s *Scene, id NodeID) {
		result.Unbind(id, true)
	})
	return result
}

// BindingSet represents a set of bindings for a specific object type.
type BindingSet[T BindingObject] struct {
	scene     *Scene
	binding   Binding[T]
	relations map[NodeID]T
}

// Bind binds the object to the node with the given ID.
func (s *BindingSet[T]) Bind(id NodeID, obj T) {
	if s.scene.IsValidNode(id) {
		s.relations[id] = obj
	}
}

// Unbind unbinds the object from its node.
func (s *BindingSet[T]) Unbind(id NodeID, notify bool) (T, bool) {
	target, exists := s.relations[id]
	if !exists {
		return gog.Zero[T](), false
	}
	if notify {
		s.binding.OnStaleBinding(s.scene, target)
	}
	return target, true
}

// Get returns the object bound to the node with the given ID. If one is
// not found, the zero value is returned.
func (s *BindingSet[T]) Get(id NodeID) T {
	return s.relations[id]
}
