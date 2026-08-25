package hierarchy

type BindingSolver[T any] interface{}

type LifecycleBindingSolver[T any] interface {
	OnDelete(*Scene, NodeID, T)
}

type SourceBindingSolver[T any] interface {
	OnSourceToNode(*Scene, NodeID, T)
}

type TargetBindingSolver[T any] interface {
	OnTargetFromNode(*Scene, NodeID, T)
}

type InterpolationBindingSolver[T any] interface {
	OnInterpolationFromNode(*Scene, NodeID, T, float64)
}

type Binding[T any] struct {
	scene               *Scene
	lifecycleSolver     LifecycleBindingSolver[T]
	sourceSolver        SourceBindingSolver[T]
	targetSolver        TargetBindingSolver[T]
	interpolationSolver InterpolationBindingSolver[T]
	bindings            map[NodeID]T
	priority            int
}

var _ internalBinding = (*Binding[any])(nil)

func NewBinding[T any](scene *Scene, solver BindingSolver[T]) *Binding[T] {
	lifecycleSolver, _ := solver.(LifecycleBindingSolver[T])
	sourceSolver, _ := solver.(SourceBindingSolver[T])
	targetSolver, _ := solver.(TargetBindingSolver[T])
	interpolationSolver, _ := solver.(InterpolationBindingSolver[T])

	result := &Binding[T]{
		scene:               scene,
		lifecycleSolver:     lifecycleSolver,
		sourceSolver:        sourceSolver,
		targetSolver:        targetSolver,
		interpolationSolver: interpolationSolver,
		bindings:            make(map[NodeID]T),
		priority:            0,
	}
	scene.addBindingSet(result)
	return result
}

func (b *Binding[T]) Delete() {
	b.scene.removeBindingSet(b)
}

func (b *Binding[T]) Priority() int {
	return b.priority
}

func (b *Binding[T]) SetPriority(priority int) {
	b.priority = priority
	b.scene.sortBindingSets()
}

func (b *Binding[T]) Bind(id NodeID, value T) {
	if !b.scene.Nodes().IsValid(id) {
		panic("cannot bind to an invalid node")
	}

	b.bindings[id] = value
}

func (b *Binding[T]) Unbind(id NodeID, notify bool) {
	if !b.scene.Nodes().IsValid(id) {
		panic("cannot unbind from an invalid node")
	}

	delete(b.bindings, id)
}

func (b *Binding[T]) Get(id NodeID) T {
	if !b.scene.Nodes().IsValid(id) {
		panic("cannot get from an invalid node")
	}

	return b.bindings[id]
}

func (b *Binding[T]) Has(id NodeID) bool {
	_, exists := b.bindings[id]
	return exists
}

func (b *Binding[T]) ApplySourcesToNodes() {
	if b.sourceSolver == nil {
		return
	}
	for id, value := range b.bindings {
		b.sourceSolver.OnSourceToNode(b.scene, id, value)
	}
}

func (b *Binding[T]) ApplySourceToNode(id NodeID) {
	if b.sourceSolver == nil {
		return
	}
	if value, exists := b.bindings[id]; exists {
		b.sourceSolver.OnSourceToNode(b.scene, id, value)
	}
}

func (b *Binding[T]) ApplyTargetsFromNodes() {
	if b.targetSolver == nil {
		return
	}
	for id, value := range b.bindings {
		b.targetSolver.OnTargetFromNode(b.scene, id, value)
	}
}

func (b *Binding[T]) ApplyTargetFromNode(id NodeID) {
	if b.targetSolver == nil {
		return
	}
	if value, exists := b.bindings[id]; exists {
		b.targetSolver.OnTargetFromNode(b.scene, id, value)
	}
}

func (b *Binding[T]) ApplyInterpolationsFromNodes(fraction float64) {
	if b.interpolationSolver == nil {
		return
	}
	for id, value := range b.bindings {
		b.interpolationSolver.OnInterpolationFromNode(b.scene, id, value, fraction)
	}
}

func (b *Binding[T]) ApplyInterpolationFromNode(id NodeID, fraction float64) {
	if b.interpolationSolver == nil {
		return
	}
	if value, exists := b.bindings[id]; exists {
		b.interpolationSolver.OnInterpolationFromNode(b.scene, id, value, fraction)
	}
}

func (b *Binding[T]) handleNodeDelete(id NodeID) {
	if b.lifecycleSolver == nil {
		return
	}
	if value, exists := b.bindings[id]; exists {
		delete(b.bindings, id)
		b.lifecycleSolver.OnDelete(b.scene, id, value)
	}
}

type internalBinding interface {
	Priority() int
	ApplySourcesToNodes()
	ApplySourceToNode(id NodeID)
	ApplyTargetsFromNodes()
	ApplyTargetFromNode(id NodeID)
	ApplyInterpolationsFromNodes(fraction float64)
	ApplyInterpolationFromNode(id NodeID, fraction float64)
	handleNodeDelete(id NodeID)
}
