package component

// State represents a persistent state of a component. Every render
// operation for a component would return the same sequence of states.
type State[T any] struct {
	node  *componentNode
	value T
	dirty bool
}

// Set changes the value stored in this State. Using this function
// will force the component to be scheduled for reconciliation.
func (s *State[T]) Set(value T) {
	s.value = value
	s.dirty = true
	// TODO: Optimize by grouping all such calls within co framework.
	uiCtx.Schedule(func() {
		if s.node.isValid() {
			s.node.reconcile(s.node.instance, s.node.scope)
		}
	})
}

// Ge returns the current value stored in this State.
func (s *State[T]) Get() T {
	return s.value
}

func (s *State[_]) isDirty() bool {
	return s.dirty
}

func (s *State[_]) setDirty(value bool) {
	s.dirty = value
}

// UseState registers a new State object to the given component.
//
// During component initialization, the closure function will be called
// to retrieve an initial value to be assigned to the State.
//
// The order in which this function is called inside a component's render
// function is important. As such, every component render should issue
// exactly the same UseState calls and in the exacly the same order.
func UseState[T any](fn func() T) *State[T] {
	if renderCtx.firstRender {
		renderCtx.node.states[renderCtx.stateDepth] = append(renderCtx.node.states[renderCtx.stateDepth], &State[T]{
			node:  renderCtx.node,
			value: fn(),
		})
	}
	result := renderCtx.node.states[renderCtx.stateDepth][renderCtx.stateIndex].(*State[T])
	renderCtx.stateIndex++
	return result
}
