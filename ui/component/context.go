package component

import (
	"fmt"
	"reflect"
)

// TODO: Add debug logging

var contexts map[reflect.Type]any

func init() {
	contexts = make(map[reflect.Type]any)
}

// RegisterContext registers a data structure that will be
// accessible from all components.
//
// If used, this method should be called during bootstrapping
// and should not be called from within components.
//
// The context is stored according to its type and there can
// only be one call per struct type. Once a context is set
// it is persisted for the whole lifecycle of the framework.
//
// Contexts should be used only for global configurations
// that will not change, like graphics handles or i18n
// functions.
func RegisterContext(value any) {
	valueType := reflect.TypeOf(value)
	if _, ok := contexts[valueType]; ok {
		panic(fmt.Errorf("a context of the specified type (%T) has already been registered", value))
	}
	contexts[valueType] = value
}

// GetContext retrieves the appropriate context based on the generic type param
// and returns it.
func GetContext[T any]() T {
	var result T
	contextValue, ok := contexts[reflect.TypeOf(result)]
	if !ok {
		panic(fmt.Errorf("there is no context of type %T", result))
	}
	return contextValue.(T)
}

// InjectContext retrieves the appropriate context and
// assigns it to target.
//
// The specified target must be a pointer to the type
// that was used in RegisterContext.
func InjectContext[T any](target *T) {
	if target == nil {
		panic("target pointer cannot be nil")
	}
	contextValue, ok := contexts[reflect.TypeOf(*target)]
	if !ok {
		panic(fmt.Errorf("there is no context of type %T", *target))
	}
	*target = contextValue.(T)
}
