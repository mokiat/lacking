package math

import gomath "math"

const Pi32 = float32(gomath.Pi)

const Epsilon32 = float32(0.000001)

func Abs32(value float32) float32 {
	return gomath.Float32frombits(gomath.Float32bits(value) &^ (1 << 31))
}

func Eq32(a, b float32) bool {
	return EqEps32(a, b, Epsilon32)
}

func EqEps32(a, b, epsilon float32) bool {
	return Abs32(a-b) < epsilon
}

func Sqrt32(value float32) float32 {
	return float32(gomath.Sqrt(float64(value)))
}
