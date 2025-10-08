package hierarchy2

const (
	initialBindingCapacity = 64
)

type BindingTarget interface {
	comparable
}

type BindingApplyFunc[T BindingTarget] func(*Scene, T, NodeID)

func NewBindingSet[T BindingTarget](scene *Scene, applyFunc BindingApplyFunc[T]) *BindingSet[T] {
	return &BindingSet[T]{
		scene:     scene,
		bindings:  make(map[T]NodeID, initialBindingCapacity),
		applyFunc: applyFunc,
	}
}

type BindingSet[T BindingTarget] struct {
	scene     *Scene
	bindings  map[T]NodeID
	applyFunc BindingApplyFunc[T]
}

func (s *BindingSet[T]) Bind(target T, id NodeID) {
	s.bindings[target] = id
}

func (s *BindingSet[T]) Unbind(target T) {
	delete(s.bindings, target)
}

func (s *BindingSet[T]) Apply() {
	for target, id := range s.bindings {
		if s.scene.IsValidNode(id) {
			s.applyFunc(s.scene, target, id)
		} else {
			delete(s.bindings, target)
		}
	}
}
