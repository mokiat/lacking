package dsl

import (
	"reflect"

	"github.com/mokiat/gomath/dprec"
)

// Degrees returns an angle constant.
func Degrees(value float64) Provider[dprec.Angle] {
	return Const(dprec.Degrees(value))
}

// Color returns a color constant.
func Color(r, g, b float64) Provider[dprec.Vec3] {
	return Const(dprec.NewVec3(r, g, b))
}

// StaticProvider creates a provider that always returns the same value.
func Const[T any](value T) Provider[T] {
	typeName := reflect.TypeOf(value).Name()
	return OnceProvider(FuncProvider(
		func() (T, error) {
			return value, nil
		},

		func() ([]byte, error) {
			return CreateDigest("const", typeName, value)
		},
	))
}
