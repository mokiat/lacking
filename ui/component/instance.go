package component

// Instance represents the instantiation of a given Component chain.
type Instance struct {
	key           string
	id            string
	name          string
	component     Component
	scopeModifier ScopeModifier
	properties    Properties
}

// Key returns the child key that is registered for this Instance
// in case the Instance was created as part of a WithChild directive.
func (i Instance) Key() string {
	return i.key
}

// Properties returns the properties assigned to this instance.
func (i *Instance) Properties() Properties {
	return i.properties
}

func (i *Instance) setID(id string) {
	i.id = id
}

func (i *Instance) setName(name string) {
	i.name = name
}

func (i *Instance) addScopeModifier(modifier ScopeModifier) {
	i.scopeModifier = ChainScopeModifier(i.scopeModifier, modifier)
}

func (i *Instance) applyScopeModifier(scope Scope) Scope {
	if i.scopeModifier == nil {
		return scope
	}
	return i.scopeModifier(scope)
}

func (i *Instance) setData(data any) {
	i.properties.data = data
}

func (i *Instance) setLayoutData(layoutData any) {
	i.properties.layoutData = layoutData
}

func (i *Instance) setCallbackData(callbackData any) {
	i.properties.callbackData = callbackData
}

func (i *Instance) setChildren(children []Instance) {
	i.properties.children = children
}

func (i *Instance) addChild(child Instance) {
	i.properties.children = append(i.properties.children, child)
}

func (i *Instance) hasMatchingChild(instance Instance) bool {
	for _, child := range i.properties.children {
		if child.key == instance.key && child.component == instance.component {
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
		component: component,
	}
	if setupFn != nil {
		setupFn()
	}
	return instanceCtx.instance
}

// WithID assigns an ID to the element of the component instance.
func WithID(id string) {
	instanceCtx.instance.setID(id)
}

// XWithID is a helper function that resembles WithID but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithID(id string) {}

// WithName assigns a name to the component instance. This is useful
// for debugging purposes.
func WithName(name string) {
	instanceCtx.instance.setName(name)
}

// XWithName is a helper function that resembles WithName but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithName(name string) {}

// WithScopeValue modifies the scope to be passed to the component instance
// by assigning the specified value to the specified key.
func WithScopeValue(key, value any) {
	instanceCtx.instance.addScopeModifier(func(scope Scope) Scope {
		return ValueScope(scope, key, value)
	})
}

// XWithScopeValue is a helper function that resembles WithScopeValue
// but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithScopeValue(key, value any) {}

// WithTypedScopeValue is a helper function that adds a value to the
// scope using the value's type as the key.
func WithTypedScopeValue[T any](value T) {
	instanceCtx.instance.addScopeModifier(func(scope Scope) Scope {
		return TypedValueScope(scope, value)
	})
}

// XWithTypedScopeValue is a helper function that resembles WithTypedScopeValue
// but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithTypedScopeValue[T any](value T) {}

// WithData specifies the data to be passed to the component instance
// during instantiation.
//
// Your data should be comparable in order to enable optimizations
// done by the framework. If you'd like to pass functions, in case of
// callbacks, they can be passed through the callback data.
func WithData(data any) {
	instanceCtx.instance.setData(data)
}

// XWithData is a helper function that resembles WithData but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithData(data any) {}

// WithLayoutData specifies the layout data to be passed to the
// component instance during instantiation.
//
// LayoutData is kept separate by the framework as it is expected
// to have a different lifecycle (changes might be rare) and as such
// can be optimized.
//
// Your layout data should be comparable in order to enable optimizations
// done by the framework.
func WithLayoutData(layoutData any) {
	instanceCtx.instance.setLayoutData(layoutData)
}

// XWithLayoutData is a helper function that resembles WithLayoutData
// but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithLayoutData(layoutData any) {}

// WithCallbackData specifies the callback data to be passed to the
// component instance during instantiation.
//
// Callback data is a mechanism for one component to listen for
// events on instanced components.
//
// As callback data is expected to be a struct of function fields,
// they are not comparable in Go and as such cannot follow the
// lifecycle of data or layout data.
func WithCallbackData(callbackData any) {
	instanceCtx.instance.setCallbackData(callbackData)
}

// XWithCallbackData is a helper function that resembles WithCallbackData
// but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithCallbackData(callbackData any) {}

// WithChild adds a child to the given component instance. The child is appended
// to all previously registered children via the same method.
//
// The key property is important. If in a subsequent render a component's
// child changes key or component type, the old one will be destroyed
// and a new one will be created. As such, to maintain a more optimized
// rendering and to prevent state loss, children should have a key assigned
// to them.
func WithChild(key string, instance Instance) {
	instance.key = key
	instanceCtx.instance.addChild(instance)
}

// XWithChild is a helper function that resembles WithChild but does nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithChild(key string, instance Instance) {}

// WithChildren sets the children for the given component instance. Keep in
// mind that any former children assigned via WithChild are replaced.
func WithChildren(children []Instance) {
	instanceCtx.instance.setChildren(children)
}

// XWithChildren is a helper function that resembles WithChildren but does
// nothing.
// This allows one to experiment during development without having to comment
// out large sections of code and deal with compilation issues.
func XWithChildren(children []Instance) {}

var instanceCtx = &instanceContext{}

type instanceContext struct {
	parent   *instanceContext
	instance Instance
}
