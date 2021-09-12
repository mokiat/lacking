package component

import (
	"fmt"
	"reflect"
)

// State represents a persistent state of a component. Every render
// operation for a component would return the same sequence of states.
type State struct {
	node  *componentNode
	value interface{}
	dirty bool
}

// Set changes the value stored in this State. Using this function
// will force the component to be scheduled for reconciliation.
func (s *State) Set(value interface{}) {
	s.value = value
	s.dirty = true
	uiCtx.Schedule(func() {
		if s.node.isValid() {
			s.node.reconcile(s.node.instance)
		}
	})
}

// Ge returns the current value stored in this State.
func (s *State) Get() interface{} {
	return s.value
}

// Inject is a helper function that can be used to inject the value of
// this state to a variable of the correct type. The specified target
// needs to be a pointer to the type of the value that was stored.
func (s *State) Inject(target interface{}) *State {
	if target == nil {
		panic("target cannot be nil")
	}
	value := reflect.ValueOf(target)
	valueType := value.Type()
	if valueType.Kind() != reflect.Ptr {
		panic("target must be a pointer")
	}
	if value.IsNil() {
		panic("target pointer cannot be nil")
	}
	stateType := reflect.TypeOf(s.value)
	if !stateType.AssignableTo(valueType.Elem()) {
		panic(fmt.Errorf("cannot assign state %T to specified type %T", s.value, target))
	}
	value.Elem().Set(reflect.ValueOf(s.value))
	return s
}

// UseState registers a new State object to the given component.
//
// During component initialization, the closure function will be called
// to retrieve an initial value to be assigned to the state.
//
// The order in which this function is called inside a component's render
// function is important. As such, every component render should issue
// exactly the same UseState calls and in the exacly the same order.
func UseState(fn func() interface{}) *State {
	if renderCtx.firstRender {
		renderCtx.node.states[renderCtx.stateDepth] = append(renderCtx.node.states[renderCtx.stateDepth], State{
			node:  renderCtx.node,
			value: fn(),
		})
	}
	result := &renderCtx.node.states[renderCtx.stateDepth][renderCtx.stateIndex]
	renderCtx.stateIndex++
	return result
}
