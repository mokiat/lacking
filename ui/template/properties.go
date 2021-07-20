package template

import (
	"fmt"
	"reflect"
)

// Properties is a holder for all data necessary to render a
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
	dataType := reflect.TypeOf(p.data)
	if !dataType.AssignableTo(valueType.Elem()) {
		panic(fmt.Errorf("cannot assign data %T to specified type %s", p.data, valueType.Elem()))
	}
	value.Elem().Set(reflect.ValueOf(p.data))
}

// LayoutData returns the layout data needed to layout the component.
func (p Properties) LayoutData() interface{} {
	return p.layoutData
}

// InjectLayoutData is a helper function that injects the LayoutData into the
// specified target, which should be a pointer to the correct type.
func (p Properties) InjectLayoutData(target interface{}) {
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
	layoutDataType := reflect.TypeOf(p.layoutData)
	if !layoutDataType.AssignableTo(valueType.Elem()) {
		panic(fmt.Errorf("cannot assign layout data %T to specified type %s", p.data, valueType.Elem()))
	}
	value.Elem().Set(reflect.ValueOf(p.layoutData))
}

// CallbackData returns the callback data that can be used by the component
// to notify its instantiator regarding key events.
func (p Properties) CallbackData() interface{} {
	return p.callbackData
}

// InjectCallbackData is a helper function that injects the CallbackData into
// the specified target, which should be a pointer to the correct type.
func (p Properties) InjectCallbackData(target interface{}) {
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
	callbackDataType := reflect.TypeOf(p.callbackData)
	if !callbackDataType.AssignableTo(valueType.Elem()) {
		panic(fmt.Errorf("cannot assign callback data %T to specified type %s", p.data, valueType.Elem()))
	}
	value.Elem().Set(reflect.ValueOf(p.callbackData))
}

// Children returns all the child instances that this component should
// host.
func (p Properties) Children() []Instance {
	return p.children
}
