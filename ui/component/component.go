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
