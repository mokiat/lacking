package template

import "reflect"

type Properties struct {
	data         interface{}
	layoutData   interface{}
	callbackData interface{}
	children     []Instance
}

func (p Properties) Data() interface{} {
	return p.data
}

func (p Properties) InjectData(target interface{}) {
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
	dataType := reflect.TypeOf(p.data)
	if !dataType.AssignableTo(valueType.Elem()) {
		panic("cannot assign data to specified type")
	}
	value.Elem().Set(reflect.ValueOf(p.data))
}

func (p Properties) LayoutData() interface{} {
	return p.layoutData
}

func (p Properties) InjectLayoutData(target interface{}) {
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
	layoutDataType := reflect.TypeOf(p.layoutData)
	if !layoutDataType.AssignableTo(valueType.Elem()) {
		panic("cannot assign layout data to specified type")
	}
	value.Elem().Set(reflect.ValueOf(p.layoutData))
}

func (p Properties) CallbackData() interface{} {
	return p.callbackData
}

func (p Properties) InjectCallbackData(target interface{}) {
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
	callbackDataType := reflect.TypeOf(p.callbackData)
	if !callbackDataType.AssignableTo(valueType.Elem()) {
		panic("cannot assign callback data to specified type")
	}
	value.Elem().Set(reflect.ValueOf(p.callbackData))
}

func (p Properties) Children() []Instance {
	return p.children
}
