package hierarchy

// BindingSolver is the type of the solver accepted by [NewBinding].
//
// A solver customizes how a [Binding] transfers data between nodes and the
// values bound to them. It is expected to implement one or more of
// [LifecycleBindingSolver], [SourceBindingSolver], [TargetBindingSolver], and
// [InterpolationBindingSolver]; the binding detects which of these a given
// solver implements and drives only those. A solver that implements none of
// them yields an inert binding that stores values but transfers nothing.
//
// The type parameter T is the type of value bound to each node. Because the
// underlying type is unconstrained, T cannot be inferred from the solver and
// must be specified explicitly when calling [NewBinding].
type BindingSolver[T any] any

// LifecycleBindingSolver is implemented by a [BindingSolver] that needs to
// react when a bound node is deleted.
type LifecycleBindingSolver[T any] interface {
	// OnDelete is invoked when a node that has a value bound in the [Binding] is
	// deleted from the scene, allowing the bound value to be released. The node
	// is still valid for the duration of the call; its binding is removed once
	// OnDelete returns.
	OnDelete(*Scene, NodeID, T)
}

// SourceBindingSolver is implemented by a [BindingSolver] that pushes state
// from a bound value onto its node; that is, the value acts as the source and
// the node as the destination.
type SourceBindingSolver[T any] interface {
	// OnSourceToNode is invoked to update the node with the specified ID from the
	// value bound to it.
	OnSourceToNode(*Scene, NodeID, T)
}

// TargetBindingSolver is implemented by a [BindingSolver] that pulls state from
// a node into its bound value; that is, the node acts as the source and the
// value as the destination.
type TargetBindingSolver[T any] interface {
	// OnTargetFromNode is invoked to update the value bound to the node with the
	// specified ID from that node.
	OnTargetFromNode(*Scene, NodeID, T)
}

// InterpolationBindingSolver is implemented by a [BindingSolver] that pulls an
// interpolated pose from a node into its bound value.
type InterpolationBindingSolver[T any] interface {
	// OnInterpolationFromNode is invoked to update the value bound to the node
	// with the specified ID from that node's transformation, interpolated by the
	// specified fraction in the range [0.0, 1.0]. See
	// [NodeView.InterpolatedAbsoluteMatrix].
	OnInterpolationFromNode(*Scene, NodeID, T, float64)
}

// Binding associates values of type T with the nodes of a [Scene] and transfers
// data between the two according to its [BindingSolver].
//
// A binding is created with [NewBinding] and belongs to a single scene. Values
// are associated with nodes via [Binding.Bind] and removed via
// [Binding.Unbind]. The transfer passes exposed by the scene
// ([Scene.ApplySourcesToNodes], [Scene.ApplyTargetsFromNodes],
// [Scene.ApplyInterpolationsFromNodes], and their single-node variants) invoke
// the corresponding solver method for each bound node. When a scene has multiple
// bindings, they are processed in ascending [Binding.Priority] order.
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

// NewBinding creates a new [Binding] of value type T within the specified scene,
// driven by the specified solver, and registers it with the scene.
//
// The solver is inspected once for the [LifecycleBindingSolver],
// [SourceBindingSolver], [TargetBindingSolver], and [InterpolationBindingSolver]
// interfaces; only those it implements take effect. Because solver is typed as
// [BindingSolver], whose underlying type is unconstrained, T cannot be inferred
// and must be provided explicitly. The new binding starts with a priority of 0.
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
	scene.addBinding(result)
	return result
}

// Delete removes the binding from its scene, so that it no longer participates
// in any transfer pass.
//
// It does not invoke [LifecycleBindingSolver.OnDelete] for any values that are
// still bound.
func (b *Binding[T]) Delete() {
	b.scene.removeBinding(b)
}

// Priority returns the priority of the binding.
//
// Within a scene, bindings are processed in ascending priority order.
func (b *Binding[T]) Priority() int {
	return b.priority
}

// SetPriority sets the priority of the binding and reorders it relative to the
// other bindings of its scene.
//
// Within a scene, bindings are processed in ascending priority order.
func (b *Binding[T]) SetPriority(priority int) {
	b.priority = priority
	b.scene.sortBindings()
}

// Bind associates the specified value with the node identified by id.
//
// If the node already has a value bound, it is replaced; the replaced value is
// not passed to [LifecycleBindingSolver.OnDelete]. It panics if id does not
// refer to a valid node.
func (b *Binding[T]) Bind(id NodeID, value T) {
	if !b.scene.Nodes().IsValid(id) {
		panic("cannot bind to an invalid node")
	}

	b.bindings[id] = value
}

// Unbind removes the value bound to the node identified by id, if any.
//
// The removal is silent: the binding's [LifecycleBindingSolver] is not informed,
// as that is reserved for node deletion. It panics if id does not refer to a
// valid node.
func (b *Binding[T]) Unbind(id NodeID) {
	if !b.scene.Nodes().IsValid(id) {
		panic("cannot unbind from an invalid node")
	}

	delete(b.bindings, id)
}

// Get returns the value bound to the node identified by id, or the zero value
// of T if the node has no value bound.
//
// It panics if id does not refer to a valid node.
func (b *Binding[T]) Get(id NodeID) T {
	if !b.scene.Nodes().IsValid(id) {
		panic("cannot get from an invalid node")
	}

	return b.bindings[id]
}

// Has returns whether the node identified by id has a value bound.
//
// It panics if id does not refer to a valid node.
func (b *Binding[T]) Has(id NodeID) bool {
	if !b.scene.Nodes().IsValid(id) {
		panic("cannot check binding for an invalid node")
	}

	_, exists := b.bindings[id]
	return exists
}

// ApplySourcesToNodes runs the source transfer for every node bound in this
// binding, updating each node from its bound value.
//
// It is a no-op if the solver does not implement [SourceBindingSolver].
func (b *Binding[T]) ApplySourcesToNodes() {
	if b.sourceSolver == nil {
		return
	}
	for id, value := range b.bindings {
		b.sourceSolver.OnSourceToNode(b.scene, id, value)
	}
}

// ApplySourceToNode runs the source transfer for the node identified by id,
// updating it from its bound value.
//
// It is a no-op if the solver does not implement [SourceBindingSolver] or if the
// node has no value bound.
func (b *Binding[T]) ApplySourceToNode(id NodeID) {
	if b.sourceSolver == nil {
		return
	}
	if value, exists := b.bindings[id]; exists {
		b.sourceSolver.OnSourceToNode(b.scene, id, value)
	}
}

// ApplyTargetsFromNodes runs the target transfer for every node bound in this
// binding, updating each bound value from its node.
//
// It is a no-op if the solver does not implement [TargetBindingSolver].
func (b *Binding[T]) ApplyTargetsFromNodes() {
	if b.targetSolver == nil {
		return
	}
	for id, value := range b.bindings {
		b.targetSolver.OnTargetFromNode(b.scene, id, value)
	}
}

// ApplyTargetFromNode runs the target transfer for the node identified by id,
// updating its bound value from the node.
//
// It is a no-op if the solver does not implement [TargetBindingSolver] or if the
// node has no value bound.
func (b *Binding[T]) ApplyTargetFromNode(id NodeID) {
	if b.targetSolver == nil {
		return
	}
	if value, exists := b.bindings[id]; exists {
		b.targetSolver.OnTargetFromNode(b.scene, id, value)
	}
}

// ApplyInterpolationsFromNodes runs the interpolation transfer for every node
// bound in this binding, updating each bound value from its node's pose
// interpolated by the specified fraction.
//
// It is a no-op if the solver does not implement [InterpolationBindingSolver].
func (b *Binding[T]) ApplyInterpolationsFromNodes(fraction float64) {
	if b.interpolationSolver == nil {
		return
	}
	for id, value := range b.bindings {
		b.interpolationSolver.OnInterpolationFromNode(b.scene, id, value, fraction)
	}
}

// ApplyInterpolationFromNode runs the interpolation transfer for the node
// identified by id, updating its bound value from the node's pose interpolated
// by the specified fraction.
//
// It is a no-op if the solver does not implement [InterpolationBindingSolver] or
// if the node has no value bound.
func (b *Binding[T]) ApplyInterpolationFromNode(id NodeID, fraction float64) {
	if b.interpolationSolver == nil {
		return
	}
	if value, exists := b.bindings[id]; exists {
		b.interpolationSolver.OnInterpolationFromNode(b.scene, id, value, fraction)
	}
}

func (b *Binding[T]) handleNodeDelete(id NodeID) {
	if value, exists := b.bindings[id]; exists {
		if b.lifecycleSolver != nil {
			b.lifecycleSolver.OnDelete(b.scene, id, value)
		}
		delete(b.bindings, id)
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
