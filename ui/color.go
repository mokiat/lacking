package ui

func RGBA(r, g, b, a uint8) Color {
	return Color{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}

func RGB(r, g, b uint8) Color {
	return RGBA(r, g, b, 255)
}

func Black() Color {
	return RGB(0, 0, 0)
}

func Maroon() Color {
	return RGB(128, 0, 0)
}

func Green() Color {
	return RGB(0, 128, 0)
}

func Navy() Color {
	return RGB(0, 0, 128)
}

func Red() Color {
	return RGB(255, 0, 0)
}

func Lime() Color {
	return RGB(0, 255, 0)
}

func Blue() Color {
	return RGB(0, 0, 255)
}

func Purple() Color {
	return RGB(128, 0, 128)
}

func Olive() Color {
	return RGB(128, 128, 0)
}

func Teal() Color {
	return RGB(0, 128, 128)
}

func Gray() Color {
	return RGB(128, 128, 128)
}

func Silver() Color {
	return RGB(192, 192, 192)
}

func Yellow() Color {
	return RGB(255, 255, 0)
}

func Fuchsia() Color {
	return RGB(255, 0, 255)
}

func Aqua() Color {
	return RGB(0, 255, 255)
}

func White() Color {
	return RGB(255, 255, 255)
}

type Color struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func (c Color) IsTransparent() bool {
	return c.A < 255
}
