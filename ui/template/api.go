package template

import (
	"fmt"
	"runtime"
)

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
