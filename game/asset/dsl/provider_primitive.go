package dsl

import (
	"fmt"
	"reflect"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
)

// Degrees returns an angle constant.
func Degrees(value float64) Provider[dprec.Angle] {
	return Const(dprec.Degrees(value))
}

// RGB returns a color constant.
func RGB(r, g, b float64) Provider[dprec.Vec4] {
	return Const(dprec.NewVec4(r, g, b, 1.0))
}

// RGBA returns a color constant.
func RGBA(r, g, b, a float64) Provider[dprec.Vec4] {
	return Const(dprec.NewVec4(r, g, b, a))
}

// StaticProvider creates a provider that always returns the same value.
func Const[T any](value T) Provider[T] {
	typeName := reflect.TypeOf(value).Name()
	return OnceProvider(FuncProvider(
		// get function
		func() (T, error) {
			return value, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("const", typeName, value)
		},
	))
}

// DPVec4ToSPVec4 converts the double precision vec4 to single precision vec4.
func DPVec4ToSPVec4(inputProvider Provider[dprec.Vec4]) Provider[sprec.Vec4] {
	return OnceProvider(FuncProvider(
		// get function
		func() (sprec.Vec4, error) {
			input, err := inputProvider.Get()
			if err != nil {
				return sprec.Vec4{}, fmt.Errorf("error getting input: %w", err)
			}
			return dtos.Vec4(input), nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("dprec-vec4-to-sprec-vec4", inputProvider)
		},
	))
}
