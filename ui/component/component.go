package component

import (
	"fmt"
	"reflect"
)

// Component represents the definition of a component.
type Component interface {
	TypeName() string

	Allocate(scope Scope, invalidate InvalidateFunc) Renderable
	Release(ref Renderable)

	NotifyCreate(ref Renderable, properties Properties)
	NotifyUpdate(ref Renderable, properties Properties)
	NotifyDelete(ref Renderable)
}

// Renderable is a component that is implemented through a Go type.
type Renderable interface {

	// Render should produce the UI hierarchy for this component.
	Render() Instance

	setScope(scope Scope)
	setProperties(properties Properties)
	setInvalidate(invalidate InvalidateFunc)
}

// InvalidateFunc can be used to indicate that the component's data has
// been internally modified and it needs to be reconciled.
type InvalidateFunc func()

// CreateNotifiable should be implemented by component types that want
// to be notified of component creation.
type CreateNotifiable interface {

	// OnCreate is called when a component is first instantiated.
	OnCreate()
}

// UpdateNotifiable should be implemented by component types that want
// to be notified of component updates.
type UpdateNotifiable interface {

	// OnUpdate is called whenever a component's properties have changed.
	OnUpdate()
}

// UpsertNotifiable should be implemented by component types that want
// to be notified of component creations and updates.
type UpsertNotifiable interface {

	// OnUpsert is called whenever a component's properties have changed.
	OnUpsert()
}

// DeleteNotifiable should be implemented by component types that want
// to be notified of component deletion.
type DeleteNotifiable interface {

	// OnDelete is called just before a component is destroyed.
	OnDelete()
}

// Define defines a new component using the specified Go type as template.
func Define(template Renderable) Component {
	return newComponentDefinition(reflect.TypeOf(template))
}

func newComponentDefinition(reflType reflect.Type) *componentDefinition {
	if reflType.Kind() == reflect.Pointer {
		return newComponentDefinition(reflType.Elem())
	}
	return &componentDefinition{
		reflType: reflType,
		name:     fmt.Sprintf("%s.%s", reflType.PkgPath(), reflType.Name()),
	}
}

type componentDefinition struct {
	reflType reflect.Type
	name     string
}

func (d *componentDefinition) TypeName() string {
	return d.name
}

func (d *componentDefinition) Allocate(scope Scope, invalidate InvalidateFunc) Renderable {
	ref := reflect.New(d.reflType).Interface().(Renderable)
	ref.setScope(scope)
	ref.setInvalidate(invalidate)
	return ref
}

func (d *componentDefinition) Release(ref Renderable) {
	ref.setScope(nil)
	ref.setInvalidate(nil)
}

func (d *componentDefinition) NotifyCreate(ref Renderable, properties Properties) {
	ref.setProperties(properties)
	if notifiable, ok := ref.(CreateNotifiable); ok {
		notifiable.OnCreate()
	}
	if notifiable, ok := ref.(UpsertNotifiable); ok {
		notifiable.OnUpsert()
	}
}

func (d *componentDefinition) NotifyUpdate(ref Renderable, properties Properties) {
	ref.setProperties(properties)
	if notifiable, ok := ref.(UpdateNotifiable); ok {
		notifiable.OnUpdate()
	}
	if notifiable, ok := ref.(UpsertNotifiable); ok {
		notifiable.OnUpsert()
	}
}

func (d *componentDefinition) NotifyDelete(ref Renderable) {
	if notifiable, ok := ref.(DeleteNotifiable); ok {
		notifiable.OnDelete()
	}
}

type BaseComponent struct {
	scope      Scope
	properties Properties
	invalidate InvalidateFunc
}

func (c *BaseComponent) Scope() Scope {
	return c.scope
}

func (c *BaseComponent) Properties() Properties {
	return c.properties
}

func (c *BaseComponent) Invalidate() {
	c.invalidate()
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
