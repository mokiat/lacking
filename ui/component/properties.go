package component

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

// LayoutData returns the layout data needed to layout the component.
func (p Properties) LayoutData() interface{} {
	return p.layoutData
}

// CallbackData returns the callback data that can be used by the component
// to notify its owner regarding key events.
func (p Properties) CallbackData() interface{} {
	return p.callbackData
}

// Children returns all the child instances that this component should host.
func (p Properties) Children() []Instance {
	return p.children
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
