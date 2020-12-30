package ui

func RGBA(r, g, b, a float32) Color {
	return Color{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}

func RGB(r, g, b float32) Color {
	return RGBA(r, g, b, 1.0)
}

func Black() Color {
	return RGB(0.0, 0.0, 0.0)
}

func Maroon() Color {
	return RGB(0.5, 0.0, 0.0)
}

func Green() Color {
	return RGB(0.0, 0.5, 0.0)
}

func Navy() Color {
	return RGB(0.0, 0.0, 0.5)
}

func Red() Color {
	return RGB(1.0, 0.0, 0.0)
}

func Lime() Color {
	return RGB(0.0, 1.0, 0.0)
}

func Blue() Color {
	return RGB(0.0, 0.0, 1.0)
}

func Purple() Color {
	return RGB(0.5, 0.0, 0.5)
}

func Olive() Color {
	return RGB(0.5, 0.5, 0.0)
}

func Teal() Color {
	return RGB(0.0, 0.5, 0.5)
}

func Gray() Color {
	return RGB(0.5, 0.5, 0.5)
}

func Silver() Color {
	return RGB(0.75, 0.75, 0.75)
}

func Yellow() Color {
	return RGB(1.0, 1.0, 0.0)
}

func Fuchsia() Color {
	return RGB(1.0, 0.0, 1.0)
}

func Aqua() Color {
	return RGB(0.0, 1.0, 1.0)
}

func White() Color {
	return RGB(1.0, 1.0, 1.0)
}

type Color struct {
	R float32
	G float32
	B float32
	A float32
}

func (c Color) IsTransparent() bool {
	return c.A < 1.0
}
