package template

// RenderContext can be used to define a render hierarchy
type RenderContext struct {
	*Context

	instance Instance
}

// Key holds the key that is propagated for this Component from its parent.
// The root component to be rendered should use this key.
func (c RenderContext) Key() string {
	return c.instance.key
}

// Data returns the data that should be used to render this Component.
func (c RenderContext) Data() interface{} {
	return c.instance.data
}

// LayoutData returns the layout information that should be assigned to
// this Component.
func (c RenderContext) LayoutData() interface{} {
	return c.instance.layoutData
}

// CallbackData returns information on callback functions.
func (c RenderContext) CallbackData() interface{} {
	return c.instance.callbackData
}

// Children returns all the child Components that should be nested into
// the current Component.
func (c RenderContext) Children() []Instance {
	return c.instance.children
}
