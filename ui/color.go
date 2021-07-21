package ui

import (
	"strconv"
	"strings"
)

var colorMap map[string]Color

func init() {
	colorMap = make(map[string]Color)
	colorMap["black"] = Black()
	colorMap["maroon"] = Maroon()
	colorMap["green"] = Green()
	colorMap["navy"] = Navy()
	colorMap["red"] = Red()
	colorMap["lime"] = Lime()
	colorMap["blue"] = Blue()
	colorMap["purple"] = Purple()
	colorMap["olive"] = Olive()
	colorMap["teal"] = Teal()
	colorMap["gray"] = Gray()
	colorMap["grey"] = Gray()
	colorMap["silver"] = Silver()
	colorMap["yellow"] = Yellow()
	colorMap["fuchsia"] = Fuchsia()
	colorMap["aqua"] = Aqua()
	colorMap["white"] = White()
}

// KnownColorName returns whether a Color with the
// specified name is known to this package.
func KnownColorName(name string) bool {
	_, ok := colorMap[strings.ToLower(name)]
	return ok
}

// NamedColor returns the Color with the specified name.
// If a Color with the specified name is not know to this
// package, then false is returned in the boolean flag.
func NamedColor(name string) (Color, bool) {
	color, ok := colorMap[strings.ToLower(name)]
	return color, ok
}

// ParseColor attempts to parse a Color from the specified
// color pattern.
// A valid pattern is `#RRGGBBAA` or `#RRGGBB`.
func ParseColor(pattern string) (Color, bool) {
	if !strings.HasPrefix(pattern, "#") {
		return Color{}, false
	}
	subPattern := pattern[1:]
	result := Color{
		A: 255,
	}
	switch len(subPattern) {
	case 8:
		if comp, err := strconv.ParseInt(subPattern[6:8], 16, 32); err == nil {
			result.A = byte(comp)
		} else {
			return Color{}, false
		}
		fallthrough
	case 6:
		if comp, err := strconv.ParseInt(subPattern[0:2], 16, 32); err == nil {
			result.R = byte(comp)
		} else {
			return Color{}, false
		}
		if comp, err := strconv.ParseInt(subPattern[2:4], 16, 32); err == nil {
			result.G = byte(comp)
		} else {
			return Color{}, false
		}
		if comp, err := strconv.ParseInt(subPattern[4:6], 16, 32); err == nil {
			result.B = byte(comp)
		} else {
			return Color{}, false
		}
		return result, true
	default:
		return Color{}, false
	}
}

// RGBA creates a new Color off of the specified
// R (Red), G (Green), B (Blue), A (Alpha) components.
func RGBA(r, g, b, a uint8) Color {
	return Color{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}

// RGB creates a new Color off of the specified
// R (Red), G (Green), B (Blue) components. The alpha
// of the Color is set to 255 (opaque).
func RGB(r, g, b uint8) Color {
	return RGBA(r, g, b, 255)
}

// Black returns an opaque Black color.
func Black() Color {
	return RGB(0, 0, 0)
}

// Maroon returns an opaque Maroon color.
func Maroon() Color {
	return RGB(128, 0, 0)
}

// Green returns an opaque Green color.
func Green() Color {
	return RGB(0, 128, 0)
}

// Navy returns an opaque Navy color.
func Navy() Color {
	return RGB(0, 0, 128)
}

// Red returns an opaque Red color.
func Red() Color {
	return RGB(255, 0, 0)
}

// Lime returns an opaque Lime color.
func Lime() Color {
	return RGB(0, 255, 0)
}

// Blue returns an opaque Blue color.
func Blue() Color {
	return RGB(0, 0, 255)
}

// Purple returns an opaque Purple color.
func Purple() Color {
	return RGB(128, 0, 128)
}

// Olive returns an opaque Olive color.
func Olive() Color {
	return RGB(128, 128, 0)
}

// Teal returns an opaque Teal color.
func Teal() Color {
	return RGB(0, 128, 128)
}

// Gray returns an opaque Gray color.
func Gray() Color {
	return RGB(128, 128, 128)
}

// Silver returns an opaque Silver color.
func Silver() Color {
	return RGB(192, 192, 192)
}

// Yellow returns an opaque Yellow color.
func Yellow() Color {
	return RGB(255, 255, 0)
}

// Fuchsia returns an opaque Fuchsia color.
func Fuchsia() Color {
	return RGB(255, 0, 255)
}

// Aqua returns an opaque Aqua color.
func Aqua() Color {
	return RGB(0, 255, 255)
}

// White returns an opaque White color.
func White() Color {
	return RGB(255, 255, 255)
}

// Transparent returns a fully transparent color.
func Transparent() Color {
	return RGBA(0, 0, 0, 0)
}

// Color represents a 32bit color (8 bits per channel).
type Color struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

// Transparent returns whether this color is transparent.
//
// A transparent color is one that is not at all visible
// (i.e. has an alpha value equal to zero).
func (c Color) Transparent() bool {
	return c.A == 0
}

// Translucent returns whether this color is translucent.
//
// A translucent color is one that is not fully visible
// (i.e. has an alpha value smaller than the maximum).
func (c Color) Translucent() bool {
	return c.A < 255
}

// Opaque returns whether this color is opaque.
// (i.e. has an alpha with the maximum value)
func (c Color) Opaque() bool {
	return c.A == 255
}
