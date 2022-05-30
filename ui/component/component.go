package component

import (
	"fmt"
	"runtime"
)

// Component represents the definition of a component.
type Component struct {
	componentType string
	componentFunc ComponentFunc
}

// ComponentFunc is the mechanism through which components can construct their
// hierarchies based on input properties and scope.
type ComponentFunc func(props Properties, scope Scope) Instance

// Define is the mechanism through which new components can be defined.
//
// The provided component function (i.e. render function) will be called by the
// framework to initialize, reconcicle,or destroy a component instance.
func Define(fn ComponentFunc) Component {
	return Component{
		componentType: evaluateComponentType(),
		componentFunc: fn,
	}
}

func evaluateComponentType() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s#%d", file, line)
}

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
