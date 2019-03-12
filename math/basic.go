package math

import gomath "math"

const Pi = float32(gomath.Pi)

func Abs32(value float32) float32 {
	return gomath.Float32frombits(gomath.Float32bits(value) &^ (1 << 31))
}
