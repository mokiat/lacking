package component

import (
	"fmt"
	"reflect"
)

var contexts map[reflect.Type]interface{}

func init() {
	contexts = make(map[reflect.Type]interface{})
}

// RegisterContext registers a data structure that will be
// accessible from all components.
//
// If used, this method should be use during bootstrapping
// and should not be called from within components.
//
// The context is stored according to its type and there can
// be only one call per struct type. Once a context is set
// it is persisted for the whole lifecycle of the framework.
//
// Contexts should be used only for global configurations
// that will not change, like graphics handles or translation
// functions.
func RegisterContext(value interface{}) {
	valueType := reflect.TypeOf(value)
	if _, ok := contexts[valueType]; ok {
		panic(fmt.Errorf("a context of the specified type %T has already been registered", value))
	}
	contexts[valueType] = value
}

// InjectContext retrieves the appropriate context and
// assigns it to target.
//
// The specified target must be a pointer to the type
// that was used in RegisterContext.
func InjectContext(target interface{}) {
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
	targetRefType := valueType.Elem()
	contextValue, ok := contexts[targetRefType]
	if !ok {
		panic(fmt.Errorf("there is no context of type %s", targetRefType))
	}
	value.Elem().Set(reflect.ValueOf(contextValue))
}
