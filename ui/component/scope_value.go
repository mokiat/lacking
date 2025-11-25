package component

import (
	"reflect"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/ui"
)

// Value is a helper function that retrieves the value as the specified
// generic param type from the specified scope using the provided key.
//
// If there is no value with the specified key in the Scope or if the value
// is not of the correct type then the zero value for that type is returned.
func Value[T any](scope Scope, key any) T {
	value, ok := scope.Value(key).(T)
	if !ok {
		return gog.Zero[T]()
	}
	return value
}

// TypedValue returns the value in the specified scope associated with the
// generic type.
//
// If there is no value with the specified type in the Scope then the zero
// value for that type is returned.
func TypedValue[T any](scope Scope) T {
	key := reflect.TypeFor[T]()
	value, ok := scope.Value(key).(T)
	if !ok {
		return gog.Zero[T]()
	}
	return value
}

// ValueScope creates a new Scope that extends the specified parent scope
// by adding the specified key-value pair.
//
// This function would not typically be used directly. Instead, use one of
// the rendering functions to modify the scope of a component instance.
func ValueScope(parent Scope, key, value any) Scope {
	return &valueScope{
		parent: parent,
		key:    key,
		value:  value,
	}
}

// TypedValueScope returns a ValueScope that uses the value's type as the
// key.
//
// This makes it easier to store and retrieve values without having
// to define custom keys.
func TypedValueScope[T any](parent Scope, value T) Scope {
	return ValueScope(parent, reflect.TypeOf(value), value)
}

type valueScope struct {
	parent Scope
	key    any
	value  any
}

func (s *valueScope) Context() *ui.Context {
	if s.parent == nil {
		return nil
	}
	return s.parent.Context()
}

func (s *valueScope) Value(key any) any {
	if s.key == key {
		return s.value
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.Value(key)
}
