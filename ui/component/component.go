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
	// NOTE: ShallowCached does not work correctly when there is an
	// intermediate container that does not take data. The only way
	// caching can be done is if all node instances caused by a node
	// are recorded as dependencies.
	return Component{
		componentType: evaluateComponentType(),
		componentFunc: delegate.componentFunc,
	}
}

func evaluateComponentType() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s#%d", file, line)
}
