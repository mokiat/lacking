package template

import (
	"fmt"
	"reflect"
)

var (
	rootState    *ReducedState
	globalStates []*ReducedState
)

type State struct {
	node  *componentNode
	value interface{}
}

func (s *State) Inject(target interface{}) {
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
}

func (s *State) Get() interface{} {
	return s.value
}

func (s *State) Set(value interface{}) {
	s.value = value
	// TODO: Schedule componentNode for reconciliation
}

type ReducedState struct {
	value   interface{}
	reducer Reducer
}

func (s *ReducedState) Inject(target interface{}) {
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
		panic("cannot assign reduced state to specified type")
	}
	value.Elem().Set(reflect.ValueOf(s.value))
}
