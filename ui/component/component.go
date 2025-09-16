package component

import (
	"fmt"
	"reflect"

	"github.com/mokiat/lacking/ui"
)

// Component represents the definition of a component.
type Component interface {

	// TypeName returns the name of this component type.
	TypeName() string

	// Allocate allocates a new instance of this component type.
	Allocate(element *ui.Element, invalidate InvalidateFunc) Renderable

	// Release releases the specified component instance.
	Release(ref Renderable)

	// HandleCreate is called when the component is first created.
	HandleCreate(ref Renderable, scope Scope, properties Properties)

	// HandleUpdate is called when the component is about to be updated
	// with new properties.
	HandleUpdate(ref Renderable, scope Scope, properties Properties)

	// HandleDelete is called when the component is about to be deleted.
	HandleDelete(ref Renderable)
}

// initializable is an internal mechanism that allows setting up
// a component instance after it has been allocated.
type initializable interface {
	setElement(element *ui.Element)
	setScope(scope Scope)
	setProperties(properties Properties)
	setInvalidate(invalidate InvalidateFunc)
}

// Renderable is a component instance.
type Renderable interface {
	initializable

	// Render should produce the component hierarchy for this component.
	Render() Instance
}

// InvalidateFunc can be used to indicate that the component's data has
// been internally modified and it needs to be reconciled.
type InvalidateFunc func()

// CreateNotifiable should be implemented by component types that want
// to be notified of component creation.
type CreateNotifiable interface {

	// OnCreate is called when a component is first instantiated.
	//
	// When called, an Element is already assigned to the component
	// and the Scope and Properties are also set.
	//
	// The Render call has not yet been called, however, as usually a component
	// would fetch data during this call that is later usind during rendering.
	// As such, no children are yet created as well.
	OnCreate()
}

// UpdateNotifiable should be implemented by component types that want
// to be notified of component updates.
type UpdateNotifiable interface {

	// OnUpdate is called whenever a component's properties have changed.
	//
	// This method is called prior to rendering the component again.
	OnUpdate()
}

// UpsertNotifiable should be implemented by component types that want
// to be notified of both component creations and updates.
type UpsertNotifiable interface {

	// OnUpsert is called whenever a component's properties have changed.
	//
	// This method is called prior to rendering the component and updating
	// or initializing any children.
	OnUpsert()
}

// DeleteNotifiable should be implemented by component types that want
// to be notified of component deletion.
type DeleteNotifiable interface {

	// OnDelete is called just before a component is destroyed.
	//
	// No rendering is performed after this call.
	OnDelete()
}

// Define defines a new component using the specified Go type as template.
func Define[T Renderable]() Component {
	return newComponentDefinition(reflect.TypeFor[T]())
}

func newComponentDefinition(reflType reflect.Type) *componentDefinition {
	if reflType.Kind() == reflect.Pointer {
		return newComponentDefinition(reflType.Elem())
	}
	return &componentDefinition{
		reflType: reflType,
	}
}

type componentDefinition struct {
	reflType reflect.Type
}

func (d *componentDefinition) TypeName() string {
	return fmt.Sprintf("%s.%s", d.reflType.PkgPath(), d.reflType.Name())
}

func (d *componentDefinition) Allocate(element *ui.Element, invalidate InvalidateFunc) Renderable {
	value := reflect.New(d.reflType)
	ref, _ := reflect.TypeAssert[Renderable](value)
	ref.setElement(element)
	ref.setInvalidate(invalidate)
	return ref
}

func (d *componentDefinition) Release(ref Renderable) {
	ref.setScope(nil)
	ref.setInvalidate(nil)
}

func (d *componentDefinition) HandleCreate(ref Renderable, scope Scope, properties Properties) {
	ref.setScope(scope)
	ref.setProperties(properties)
	if notifiable, ok := ref.(CreateNotifiable); ok {
		notifiable.OnCreate()
	}
	if notifiable, ok := ref.(UpsertNotifiable); ok {
		notifiable.OnUpsert()
	}
}

func (d *componentDefinition) HandleUpdate(ref Renderable, scope Scope, properties Properties) {
	ref.setScope(scope)
	ref.setProperties(properties)
	if notifiable, ok := ref.(UpdateNotifiable); ok {
		notifiable.OnUpdate()
	}
	if notifiable, ok := ref.(UpsertNotifiable); ok {
		notifiable.OnUpsert()
	}
}

func (d *componentDefinition) HandleDelete(ref Renderable) {
	if notifiable, ok := ref.(DeleteNotifiable); ok {
		notifiable.OnDelete()
	}
}

// BaseComponent contains the basic functionality needed to implement a
// Renderable component. All component implementations should embed this struct.
type BaseComponent struct {
	element    *ui.Element
	scope      Scope
	properties Properties
	invalidate InvalidateFunc
}

var _ initializable = (*BaseComponent)(nil)

// Name returns the name of this component. This is mostly useful for
// debugging purposes.
func (c *BaseComponent) Name() string {
	node := componentNodeFromScope(c.scope)
	return node.name
}

// Element returns the underlying UI element of this component.
func (c *BaseComponent) Element() *ui.Element {
	return c.element
}

// Scope returns the scope in which this component is rendered.
func (c *BaseComponent) Scope() Scope {
	return c.scope
}

// Properties returns the properties with which this component was rendered.
func (c *BaseComponent) Properties() Properties {
	return c.properties
}

// Invalidate marks this component as needing to be reconciled.
func (c *BaseComponent) Invalidate() {
	c.invalidate()
}

func (c *BaseComponent) setElement(element *ui.Element) {
	c.element = element
}

func (c *BaseComponent) setScope(scope Scope) {
	c.scope = scope
}

func (c *BaseComponent) setProperties(properties Properties) {
	c.properties = properties
}

func (c *BaseComponent) setInvalidate(invalidate InvalidateFunc) {
	c.invalidate = invalidate
}
