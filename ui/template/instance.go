package template

import "github.com/mokiat/lacking/ui"

// Instance represents the instantiation of a given Component.
type Instance struct {
	key           string
	componentType string
	componentFunc ComponentFunc

	data         interface{}
	layoutData   interface{}
	callbackData interface{}
	children     []Instance

	element *ui.Element
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

// New instantiates the specified component. The setup function is used
// to apply configurations to the component in a user-friendly manner.
//
// Note: creating an instance with New does not necessarily mean that a
// component will be freshly instantiated. If this occurs during rendering
// the framework will reuse former instances when possible.
func New(component Component, setupFn func()) Instance {
	dslCtx = &dslContext{
		parent: dslCtx,
	}
	defer func() {
		dslCtx = dslCtx.parent
	}()

	dslCtx.instance = Instance{
		componentType: component.componentType,
		componentFunc: component.componentFunc,
	}
	setupFn()
	return dslCtx.instance
}

// WithContext can be used during the instantiation of an Application
// in order to configure a context object.
//
// This is a helper function in place of RegisterContext. While currently not
// enforced, you should use this function during the instantiation of your
// root component.
// Using it at a later point during the lifecycle of your application could
// indicate an improper usage of contexts. You may consider using reducers
// and global state instead.
func WithContext(context interface{}) {
	RegisterContext(context)
}

// WithData specifies the data to be passed to the component
// during instantiation.
//
// Your data should be comparable in order to enable optimizations
// done by the framework. If you'd like to pass functions, in case of
// callbacks, they can be passed through callback data.
func WithData(data interface{}) {
	dslCtx.instance.data = data
}

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
	dslCtx.instance.layoutData = layoutData
}

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
	dslCtx.instance.callbackData = callbackData
}

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
	dslCtx.instance.children = append(dslCtx.instance.children, instance)
}

// WithChildren sets the children for the given component. Keep in mind that
// any former children assigned via WithChild are replaced.
func WithChildren(children []Instance) {
	dslCtx.instance.children = children
}
