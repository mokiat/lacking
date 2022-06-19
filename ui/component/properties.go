package component

import (
	"fmt"
	"reflect"
)

// Properties is a holder for all user-specified data necessary to render a
// component.
type Properties struct {
	data         interface{}
	layoutData   interface{}
	callbackData interface{}
	children     []Instance
}

// Data returns the configuration data needed to render the component.
func (p Properties) Data() interface{} {
	return p.data
}

// InjectData is a helper function that injects the Data into the
// specified target, which should be a pointer to the correct type.
func (p Properties) InjectData(target interface{}) {
	inject(target, p.data)
}

// InjectOptionalData is a helper function that injects the Data into the
// specified target, which should be a pointer to the correct type, or if there
// is no data, it injects the default one.
func (p Properties) InjectOptionalData(target, defaultValue interface{}) {
	if p.data != nil {
		inject(target, p.data)
	} else {
		inject(target, defaultValue)
	}
}

// GetData returns the data stored in Properties as the specified type.
func GetData[T any](props Properties) T {
	return props.data.(T)
}

// GetOptionalData returns the data stored in Properties as the specified type,
// unless there is no data, in which case the defaultValue is returned.
func GetOptionalData[T any](props Properties, defaultValue T) T {
	if props.data == nil {
		return defaultValue
	}
	return props.data.(T)
}

// LayoutData returns the layout data needed to layout the component.
func (p Properties) LayoutData() interface{} {
	return p.layoutData
}

// InjectLayoutData is a helper function that injects the LayoutData into the
// specified target, which should be a pointer to the correct type.
func (p Properties) InjectLayoutData(target interface{}) {
	inject(target, p.layoutData)
}

// InjectOptionalLayoutData is a helper function that injects the LayoutData
// into the specified target, which should be a pointer to the correct type, or
// if there is no layout data, it injects the defaultValue one.
func (p Properties) InjectOptionalLayoutData(target, defaultValue interface{}) {
	if p.layoutData != nil {
		inject(target, p.layoutData)
	} else {
		inject(target, defaultValue)
	}
}

// GetLayoutData returns the layout data stored in Properties as the
// specified type.
func GetLayoutData[T any](props Properties) T {
	return props.layoutData.(T)
}

// GetOptionalLayoutData returns the layout data stored in Properties as
// the specified type, unless there is no layout data, in which case the
// defaultValue is returned.
func GetOptionalLayoutData[T any](props Properties, defaultValue T) T {
	if props.layoutData == nil {
		return defaultValue
	}
	return props.layoutData.(T)
}

// CallbackData returns the callback data that can be used by the component
// to notify its instantiator regarding key events.
func (p Properties) CallbackData() interface{} {
	return p.callbackData
}

// InjectCallbackData is a helper function that injects the CallbackData into
// the specified target, which should be a pointer to the correct type.
func (p Properties) InjectCallbackData(target interface{}) {
	inject(target, p.callbackData)
}

// InjectOptionalCallbackData is a helper function that injects the CallbackData
// into the specified target, which should be a pointer to the correct type, or
// if there is no callback data, it injects the defaultValue one.
func (p Properties) InjectOptionalCallbackData(target, defaultValue interface{}) {
	if p.callbackData != nil {
		inject(target, p.callbackData)
	} else {
		inject(target, defaultValue)
	}
}

// GetCallbackData returns the callback data stored in Properties as the
// specified type.
func GetCallbackData[T any](props Properties) T {
	return props.callbackData.(T)
}

// GetOptionalCallbackData returns the callback data stored in Properties as
// the specified type, unless there is no callback data, in which case the
// defaultValue is returned.
func GetOptionalCallbackData[T any](props Properties, defaultValue T) T {
	if props.callbackData == nil {
		return defaultValue
	}
	return props.callbackData.(T)
}

// Children returns all the child instances that this component should host.
func (p Properties) Children() []Instance {
	return p.children
}

func inject(target, injectedValue interface{}) {
	if target == nil {
		panic(fmt.Errorf("target cannot be nil"))
	}
	value := reflect.ValueOf(target)
	valueType := value.Type()
	if valueType.Kind() != reflect.Ptr {
		panic(fmt.Errorf("target %T must be a pointer", target))
	}
	if value.IsNil() {
		panic(fmt.Errorf("target pointer cannot be nil"))
	}
	callbackDataType := reflect.TypeOf(injectedValue)
	if !callbackDataType.AssignableTo(valueType.Elem()) {
		panic(fmt.Errorf("cannot assign value of type %T to specified reference type %s", injectedValue, valueType.Elem()))
	}
	value.Elem().Set(reflect.ValueOf(injectedValue))
}
