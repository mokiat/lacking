package mdl

import "github.com/x448/float16"

func RGBA8Color(r, g, b, a uint8) Color {
	return Color{
		R: float64(r) / 255.0,
		G: float64(g) / 255.0,
		B: float64(b) / 255.0,
		A: float64(a) / 255.0,
	}
}

func RGBA16FColor(r, g, b, a float16.Float16) Color {
	return Color{
		R: float64(r.Float32()),
		G: float64(g.Float32()),
		B: float64(b.Float32()),
		A: float64(a.Float32()),
	}
}

func RGBA32FColor(r, g, b, a float32) Color {
	return Color{
		R: float64(r),
		G: float64(g),
		B: float64(b),
		A: float64(a),
	}
}

func RGBA64FColor(r, g, b, a float64) Color {
	return Color{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}

type Color struct {
	R float64
	G float64
	B float64
	A float64
}

func (c Color) RGBA8() (r, g, b, a uint8) {
	return uint8(c.R * 255.0),
		uint8(c.G * 255.0),
		uint8(c.B * 255.0),
		uint8(c.A * 255.0)
}

func (c Color) RGBA16F() (r, g, b, a float16.Float16) {
	return float16.Fromfloat32(float32(c.R)),
		float16.Fromfloat32(float32(c.G)),
		float16.Fromfloat32(float32(c.B)),
		float16.Fromfloat32(float32(c.A))
}

func (c Color) RGBA32F() (r, g, b, a float32) {
	return float32(c.R),
		float32(c.G),
		float32(c.B),
		float32(c.A)
}

func (c Color) RGBA64F() (r, g, b, a float64) {
	return c.R, c.G, c.B, c.A
}
