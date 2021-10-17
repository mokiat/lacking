package component

import (
	"fmt"
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

func Controlled(delegate Component) Component {
	type controlledState struct {
		controller   Controller
		subscription ControllerSubscription
	}

	controllers := make(map[*componentNode]controlledState)

	return Component{
		componentType: evaluateComponentType(),
		componentFunc: func(props Properties) Instance {
			controller := props.Data().(Controller)
			node := renderCtx.node

			if state, ok := controllers[renderCtx.node]; ok {
				if controller != state.controller {
					state.subscription.Unsubscribe()
					state.controller = controller
					state.subscription = controller.Subscribe(func(controller Controller) {
						uiCtx.Schedule(func() {
							if node.isValid() {
								node.reconcile(node.instance)
							}
						})
					})
				}
			} else {
				controllers[node] = controlledState{
					controller: controller,
					subscription: controller.Subscribe(func(controller Controller) {
						uiCtx.Schedule(func() {
							if node.isValid() {
								node.reconcile(node.instance)
							}
						})
					}),
				}
			}
			return delegate.componentFunc(props)
		},
	}
}

// ShallowCached can be used to wrap a component and optimize
// reconciliation by avoiding the rerendering of the component
// if the data and layout data are equal to their previous values
// when shallowly (==) compared.
func ShallowCached(delegate Component) Component {
	// FIXME: Cache grows indefinitely
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
					!IsEqualData(oldData, props.data) ||
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

func evaluateComponentType() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s#%d", file, line)
}

func isLayoutDataShallowEqual(oldLayoutData, newLayoutData interface{}) bool {
	return newLayoutData == oldLayoutData
}

func areChildrenEqual(oldChildren, newChildren []Instance) bool {
	if len(newChildren) != len(oldChildren) {
		return false
	}
	for i, newChild := range newChildren {
		oldChild := oldChildren[i]
		if newChild.key != oldChild.key {
			return false
		}
		if newChild.componentType != oldChild.componentType {
			return false
		}
		if !IsEqualData(oldChild.data, newChild.data) {
			return false
		}
	}
	return true
}
