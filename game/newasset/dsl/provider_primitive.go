package dsl

import "reflect"

// Int returns an int constant.
func Int(value int) Provider[int] {
	return OnceProvider(StaticProvider(value))
}

// StaticProvider creates a provider that always returns the same value.
func StaticProvider[T any](value T) Provider[T] {
	typeName := reflect.TypeOf(value).Name()
	return OnceProvider(FuncProvider(
		func() (T, error) {
			return value, nil
		},

		func() ([]byte, error) {
			return digestItems("static", typeName, value)
		},
	))
}
