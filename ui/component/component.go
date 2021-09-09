package component

import (
	"fmt"
	"reflect"
	"runtime"
)

// Component represents a definition for a component.
type Component struct {
	componentType string
	componentFunc ComponentFunc
}

// Define can be used to describe a new component. The provided component
// function (or render function) will be called by the framework to
// initialize, reconcicle, or destroy a component instance.
func Define(fn ComponentFunc) Component {
	return Component{
		componentType: evaluateComponentType(),
		componentFunc: fn,
	}
}

// ComponentFunc holds the logic and layouting of the component.
type ComponentFunc func(props Properties) Instance

type cachedState struct {
	oldData        interface{}
	oldLayoutData  interface{}
	oldChildren    []Instance
	cachedInstance Instance
}

// ShallowCached can be used to wrap a component and optimize
// reconciliation by avoiding the rerendering of the component
// if the data and layout data are equal to their previous values
// when shallowly (==) compared.
func ShallowCached(delegate Component) Component {
	cache := make(map[*componentNode]*cachedState)

	return Component{
		componentType: evaluateComponentType(),
		componentFunc: func(props Properties) Instance {
			if state, ok := cache[renderCtx.node]; ok {
				oldData := state.oldData
				oldLayoutData := state.oldLayoutData
				oldChildren := state.oldChildren

				shouldCallDelegate := renderCtx.lastRender ||
					renderCtx.forcedRender ||
					((oldData == nil) && (oldLayoutData == nil) && (oldChildren == nil)) ||
					!isDataShallowEqual(oldData, props.data) ||
					!isLayoutDataShallowEqual(oldLayoutData, props.layoutData) ||
					!areChildrenEqual(oldChildren, props.children)
				if !shouldCallDelegate {
					return state.cachedInstance
				}
			}

			instance := delegate.componentFunc(props)
			cache[renderCtx.node] = &cachedState{
				oldData:        props.data,
				oldLayoutData:  props.layoutData,
				oldChildren:    props.children,
				cachedInstance: instance,
			}
			return instance
		},
	}
}

// DeepCached can be used to wrap a component and optimize
// reconciliation by avoiding the rerendering of the component
// if the data and layout data are equal to their previous values
// when deeply compared.
func DeepCached(delegate Component) Component {
	cache := make(map[*componentNode]*cachedState)

	return Component{
		componentType: evaluateComponentType(),
		componentFunc: func(props Properties) Instance {
			if state, ok := cache[renderCtx.node]; ok {
				oldData := state.oldData
				oldLayoutData := state.oldLayoutData
				oldChildren := state.oldChildren

				shouldCallDelegate := renderCtx.lastRender ||
					renderCtx.forcedRender ||
					((oldData == nil) && (oldLayoutData == nil) && (oldChildren == nil)) ||
					!isDataDeepEqual(oldData, props.data) ||
					!isLayoutDataDeepEqual(oldLayoutData, props.layoutData) ||
					!areChildrenEqual(oldChildren, props.children)
				if !shouldCallDelegate {
					return state.cachedInstance
				}
			}

			instance := delegate.componentFunc(props)
			cache[renderCtx.node] = &cachedState{
				oldData:        props.data,
				oldLayoutData:  props.layoutData,
				oldChildren:    props.children,
				cachedInstance: instance,
			}
			return instance
		},
	}
}

func evaluateComponentType() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s#%d", file, line)
}

func isDataShallowEqual(oldData, newData interface{}) bool {
	return newData == oldData
}

func isDataDeepEqual(oldData, newData interface{}) bool {
	return reflect.DeepEqual(newData, oldData)
}

func isLayoutDataShallowEqual(oldLayoutData, newLayoutData interface{}) bool {
	return newLayoutData == oldLayoutData
}

func isLayoutDataDeepEqual(oldLayoutData, newLayoutData interface{}) bool {
	return reflect.DeepEqual(newLayoutData, oldLayoutData)
}

func areChildrenEqual(oldChildren, newChildren []Instance) bool {
	if len(newChildren) != len(oldChildren) {
		return false
	}
	for i := range newChildren {
		if newChildren[i].key != oldChildren[i].key {
			return false
		}
		if newChildren[i].componentType != oldChildren[i].componentType {
			return false
		}
	}
	return true
}
