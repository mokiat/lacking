package template

import (
	"fmt"
	"reflect"
	"runtime"
)

var (
	rootState    *ReducedState
	globalStates []*ReducedState
)

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

func InitGlobalState(state *ReducedState) {
	rootState = state
}

func NewReducedState(reducer Reducer) *ReducedState {
	result := &ReducedState{
		reducer: reducer,
		value:   reducer(nil, nil),
	}
	globalStates = append(globalStates, result)
	return result
}

func Dispatch(action interface{}) {
	invalidateGlobalNodes := false
	for _, state := range globalStates {
		newValue := state.reducer(state, action)
		if newValue != state.value {
			state.value = newValue
			invalidateGlobalNodes = true
		}
	}
	if invalidateGlobalNodes {
		for _, node := range globalStateNodes {
			node.reconcile(node.instance)
		}
	}
}

var globalStateNodes []*componentNode

type Reducer func(state *ReducedState, action interface{}) interface{}

type ConnectFunc func(props Properties, rootState *ReducedState) (data interface{}, callbackData interface{})

func Connect(delegate Component, connectFn ConnectFunc) Component {
	_, file, line, _ := runtime.Caller(1)
	return Component{
		componentType: fmt.Sprintf("%s#%d", file, line),
		componentFunc: func(props Properties) Instance {
			Once(func() {
				globalStateNodes = append(globalStateNodes, renderCtx.node)
			})

			Defer(func() {
				for i, node := range globalStateNodes {
					if node == renderCtx.node {
						globalStateNodes[i] = globalStateNodes[len(globalStateNodes)-1]
						globalStateNodes = globalStateNodes[:len(globalStateNodes)-1]
					}
				}
			})

			data, callbackData := connectFn(props, rootState)
			return delegate.componentFunc(Properties{
				data:         data,
				layoutData:   props.layoutData,
				callbackData: callbackData,
				children:     props.children,
			})
		},
	}
}
