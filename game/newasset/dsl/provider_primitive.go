package dsl

import (
	"reflect"

	"github.com/mokiat/gomath/dprec"
)

// Int returns an int constant.
func Int(value int) Provider[int] {
	return OnceProvider(StaticProvider(value))
}

// Float returns a float constant.
func Float(value float64) Provider[float64] {
	return OnceProvider(StaticProvider(value))
}

// Degrees returns an angle constant.
func Degrees(value float64) Provider[dprec.Angle] {
	return OnceProvider(StaticProvider(dprec.Degrees(value)))
}

// Color returns a color constant.
func Color(r, g, b float64) Provider[dprec.Vec3] {
	return OnceProvider(StaticProvider(dprec.NewVec3(r, g, b)))
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
