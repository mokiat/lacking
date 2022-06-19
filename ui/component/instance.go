package component

import "github.com/mokiat/lacking/ui"

var instanceCtx = &instanceContext{}

type instanceContext struct {
	parent   *instanceContext
	instance Instance
}

// Instance represents the instantiation of a given Component.
type Instance struct {
	key           string
	componentType string
	componentFunc ComponentFunc

	data         interface{}
	layoutData   interface{}
	callbackData interface{}
	scope        Scope
	children     []Instance

	element *ui.Element
}

// Key returns the child key that is registered for this Instance
// in case the Instance was created as part of a WithChild directive.
func (i Instance) Key() string {
	return i.key
}

func (i Instance) properties() Properties {
	return Properties{
		data:         i.data,
		layoutData:   i.layoutData,
		callbackData: i.callbackData,
		children:     i.children,
	}
}

func (i Instance) hasMatchingChild(instance Instance) bool {
	for _, child := range i.children {
		if child.key == instance.key && child.componentType == instance.componentType {
			return true
		}
	}
	return false
}

// New instantiates the specified component. The setupFn function is used
// to apply configurations to the component in a DSL manner.
//
// NOTE: Creating an instance with New does not necessarily mean that a
// component will be freshly instantiated. If this occurs during re-rendering
// the framework will reuse former instances when possible.
func New(component Component, setupFn func()) Instance {
	instanceCtx = &instanceContext{
		parent: instanceCtx,
	}
	defer func() {
		instanceCtx = instanceCtx.parent
	}()

	instanceCtx.instance = Instance{
		componentType: component.componentType,
		componentFunc: component.componentFunc,
	}
	if setupFn != nil {
		setupFn()
	}
	return instanceCtx.instance
}

// WithData specifies the data to be passed to the component
// during instantiation.
//
// Your data should be comparable in order to enable optimizations
// done by the framework. If you'd like to pass functions, in case of
// callbacks, they can be passed through the callback data.
func WithData(data interface{}) {
	instanceCtx.instance.data = data
}

// XWithData is a helper function that resembles WithData but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithData(data interface{}) {}

// WithLayoutData specifies the layout data to be passed to the
// component during instantiation.
//
// LayoutData is kept separate by the framework as it is expected
// to have a different lifecycle (changes might be rare) and as such
// can be optimized.
//
// Your layout data should be comparable in order to enable optimizations
// done by the framework.
func WithLayoutData(layoutData interface{}) {
	instanceCtx.instance.layoutData = layoutData
}

// XWithLayoutData is a helper function that resembles WithLayoutData
// but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithLayoutData(layoutData interface{}) {}

// WithCallbackData specifies the callback data to be passed to the
// component during instantiation.
//
// Callback data is a mechanism for one component to listen for
// events on instanced components.
//
// As callback data is expected to be a struct of function fields,
// they are not comparable in Go and as such cannot follow the
// lifecycle of data or layout data.
func WithCallbackData(callbackData interface{}) {
	instanceCtx.instance.callbackData = callbackData
}

// XWithCallbackData is a helper function that resembles WithCallbackData
// but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithCallbackData(callbackData interface{}) {}

// WithScope attaches a custom Scope to this component. Any child components
// will inherit the Scope unless overridden by another call to WithScope.
func WithScope(scope Scope) {
	instanceCtx.instance.scope = scope
}

// XWithScope is a helper function that resembles WithScope but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithScope(scope Scope) {}

// WithChild adds a child to the given component. The child is appended
// to all previously registered children via the same method.
//
// The key property is important. If in a subsequent render a component's
// child changes key or component type, the old one will be destroyed
// and a new one will be created. As such, to maintain a more optimized
// rendering and to prevent state loss, children should have a key assigned
// to them.
func WithChild(key string, instance Instance) {
	instance.key = key
	instanceCtx.instance.children = append(instanceCtx.instance.children, instance)
}

// XWithChild is a helper function that resembles WithChild but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithChild(key string, instance Instance) {}

// WithChildren sets the children for the given component. Keep in mind that
// any former children assigned via WithChild are replaced.
func WithChildren(children []Instance) {
	instanceCtx.instance.children = children
}

// XWithChildren is a helper function that resembles WithChildren but does
// nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithChildren(children []Instance) {}
