package component

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/mokiat/lacking/log"
)

// Component represents the definition of a component.
type Component struct {
	componentType string
	componentFunc ComponentFunc
}

// ComponentFunc is the mechanism through which components can construct their
// hierarchies based on input properties and scope.
//
// Deprecated: Use DefineType instead
type ComponentFunc func(props Properties, scope Scope) Instance

// Define is the mechanism through which new components can be defined.
//
// The provided component function (i.e. render function) will be called by the
// framework to initialize, reconcicle,or destroy a component instance.
//
// Deprecated: Use DefineType instead
func Define(fn ComponentFunc) Component {
	return Component{
		componentType: evaluateComponentType(),
		componentFunc: fn,
	}
}

func evaluateComponentType() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s#%d", file, line)
}

// Renderable is a component that is implemented through a Go type.
type Renderable interface {

	// Render should produce the UI hierarchy for this component.
	Render() Instance
}

// CreateNotifiable should be implemented by TypeComponent types that want
// to be notified of component creation.
type CreateNotifiable interface {

	// OnCreate is called when a component is first instantiated.
	OnCreate()
}

// UpdateNotifiable should be implemented by TypeComponent types that want
// to be notified of component update.
type UpdateNotifiable interface {

	// OnUpdate is called whenever a component's properties have changed.
	OnUpdate()
}

// UpsertNotifiable should be implemented by TypeComponent types that want
// to be notified of component changes.
type UpsertNotifiable interface {

	// OnUpsert is called when the component is first instantiated and whenever
	// the properties have changed.
	OnUpsert()
}

// DeleteNotifiable should be implemented by TypeComponent types that want
// to be notified of component deletion.
type DeleteNotifiable interface {

	// OnDelete is called just before a component is destroyed.
	OnDelete()
}

// DefineType defines a new component using the specified Go type as template.
func DefineType(template Renderable) Component {
	// TODO: Flip things around. Have this be the main way to create
	// components and reimplement the old behavior to use this internally
	// for storing state and notifications.

	definition := newTypeComponentDefinition(reflect.TypeOf(template))

	return Component{
		componentType: definition.Name(),
		componentFunc: func(props Properties, scope Scope) Instance {
			presenter := UseState(func() Renderable {
				// TODO: Consider instantiating the ui Element here and assign it to the
				// instance, if the instance has an `element` tag. That way the element
				// will be available from the beginning, even if not initially attached.

				return definition.NewInstance()
			}).Get()

			invalidateProp := UseState(func() int {
				return 0
			})

			invalidate := UseState(func() func() {
				return func() {
					invalidateProp.Set(invalidateProp.Get() + 1)
				}
			}).Get()

			var justCreated bool
			Once(func() {
				justCreated = true
				target := reflect.ValueOf(presenter).Elem()
				definition.AssignProperties(target, invalidate, scope, props)
				if notifiable, ok := presenter.(CreateNotifiable); ok {
					notifiable.OnCreate()
				}
				if notifiable, ok := presenter.(UpsertNotifiable); ok {
					notifiable.OnUpsert()
				}
			})

			var justDeleted bool
			Defer(func() {
				justDeleted = true
				if notifiable, ok := presenter.(DeleteNotifiable); ok {
					notifiable.OnDelete()
				}
			})

			if !justCreated && !justDeleted {
				target := reflect.ValueOf(presenter).Elem()
				definition.AssignProperties(target, invalidate, scope, props)
				if notifiable, ok := presenter.(UpdateNotifiable); ok {
					notifiable.OnUpdate()
				}
				if notifiable, ok := presenter.(UpsertNotifiable); ok {
					notifiable.OnUpsert()
				}
			}

			return presenter.Render()
		},
	}
}

func newTypeComponentDefinition(reflType reflect.Type) typeComponentDefinition {
	if reflType.Kind() == reflect.Pointer {
		return newTypeComponentDefinition(reflType.Elem())
	}

	var (
		scopeFieldIndices        []int
		propertiesFieldIndices   []int
		dataFieldIndices         []int
		callbackDataFieldIndices []int
		childrenFieldIndices     []int
		layoutDataFieldIndices   []int
		invalidateFieldIndices   []int
	)

	if reflType.Kind() == reflect.Struct {
		for i := 0; i < reflType.NumField(); i++ {
			field := reflType.Field(i)
			switch tag := field.Tag.Get("co"); tag {
			case "":
				// no tag
			case "props", "properties":
				propertiesFieldIndices = append(propertiesFieldIndices, i)
			case "scope":
				scopeFieldIndices = append(scopeFieldIndices, i)
			case "data":
				dataFieldIndices = append(dataFieldIndices, i)
			case "callback":
				callbackDataFieldIndices = append(callbackDataFieldIndices, i)
			case "children":
				childrenFieldIndices = append(childrenFieldIndices, i)
			case "layout":
				layoutDataFieldIndices = append(layoutDataFieldIndices, i)
			case "invalidate":
				invalidateFieldIndices = append(invalidateFieldIndices, i)
			default:
				log.Warn("Unknown type component field tag %q!", tag)
			}
		}
	}

	return typeComponentDefinition{
		reflType:                 reflType,
		name:                     fmt.Sprintf("%s.%s", reflType.PkgPath(), reflType.Name()),
		scopeFieldIndices:        scopeFieldIndices,
		propertiesFieldIndices:   propertiesFieldIndices,
		dataFieldIndices:         dataFieldIndices,
		callbackDataFieldIndices: callbackDataFieldIndices,
		childrenFieldIndices:     childrenFieldIndices,
		layoutDataFieldIndices:   layoutDataFieldIndices,
		invalidateFieldIndices:   invalidateFieldIndices,
	}
}

type typeComponentDefinition struct {
	reflType                 reflect.Type
	name                     string
	scopeFieldIndices        []int
	propertiesFieldIndices   []int
	dataFieldIndices         []int
	callbackDataFieldIndices []int
	childrenFieldIndices     []int
	layoutDataFieldIndices   []int
	invalidateFieldIndices   []int
}

func (d typeComponentDefinition) Name() string {
	return d.name
}

func (d typeComponentDefinition) NewInstance() Renderable {
	return reflect.New(d.reflType).Interface().(Renderable)
}

func (d typeComponentDefinition) AssignProperties(target reflect.Value, invalidate func(), scope Scope, props Properties) {
	for _, index := range d.scopeFieldIndices {
		target.Field(index).Set(reflect.ValueOf(scope))
	}
	for _, index := range d.propertiesFieldIndices {
		target.Field(index).Set(reflect.ValueOf(props))
	}
	if value := props.data; value != nil {
		for _, index := range d.dataFieldIndices {
			target.Field(index).Set(reflect.ValueOf(value))
		}
	}
	if value := props.callbackData; value != nil {
		for _, index := range d.callbackDataFieldIndices {
			target.Field(index).Set(reflect.ValueOf(value))
		}
	}
	if value := props.layoutData; value != nil {
		for _, index := range d.layoutDataFieldIndices {
			target.Field(index).Set(reflect.ValueOf(value))
		}
	}
	for _, index := range d.childrenFieldIndices {
		target.Field(index).Set(reflect.ValueOf(props.children))
	}
	for _, index := range d.invalidateFieldIndices {
		target.Field(index).Set(reflect.ValueOf(invalidate))
	}
}
