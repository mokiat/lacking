package component

import (
	"fmt"
	"runtime"
)

// Component represents a definition for a component.
type Component struct {
	componentType string
	componentFunc ComponentFuncV2
}

// Define can be used to describe a new component. The provided component
// function (or render function) will be called by the framework to
// initialize, reconcicle, or destroy a component instance.
func Define(fn ComponentFunc) Component {
	return Component{
		componentType: evaluateComponentType(),
		componentFunc: func(props Properties, scope Scope) Instance {
			return fn(props)
		},
	}
}

// TODO
func DefineV2(fn ComponentFuncV2) Component {
	return Component{
		componentType: evaluateComponentType(),
		componentFunc: fn,
	}
}

// ComponentFunc holds the logic and layouting of the component.
type ComponentFunc func(props Properties) Instance

// TODO
type ComponentFuncV2 func(props Properties, scope Scope) Instance

// TODO: Remove. This can be achieved with mvc or lifecycles.
func Controlled(delegate Component) Component {
	type controlledState struct {
		controller   Controller
		subscription ControllerSubscription
	}

	controllers := make(map[*componentNode]controlledState)

	return Component{
		componentType: evaluateComponentType(),
		componentFunc: func(props Properties, scope Scope) Instance {
			controller := props.Data().(Controller)
			node := renderCtx.node

			if state, ok := controllers[renderCtx.node]; ok {
				if controller != state.controller {
					state.subscription.Unsubscribe()
					state.controller = controller
					state.subscription = controller.Subscribe(func(controller Controller) {
						uiCtx.Schedule(func() {
							if node.isValid() {
								node.reconcile(node.instance, node.scope)
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
								node.reconcile(node.instance, node.scope)
							}
						})
					}),
				}
			}
			return delegate.componentFunc(props, scope)
		},
	}
}

func evaluateComponentType() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s#%d", file, line)
}
