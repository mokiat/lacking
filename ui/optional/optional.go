package optional

import "github.com/mokiat/lacking/ui"

// NewBool returns a specified Bool value.
func NewBool(value bool) Bool {
	return Bool{
		Specified: true,
		Value:     value,
	}
}

// Bool represents an optional bool value.
type Bool struct {
	Specified bool
	Value     bool
}

// NewInt returns a specified Int value.
func NewInt(value int) Int {
	return Int{
		Specified: true,
		Value:     value,
	}
}

// Int represents an optional int value.
type Int struct {
	Specified bool
	Value     int
}

// NewFloat32 returns a specified Float32 value.
func NewFloat32(value float32) Float32 {
	return Float32{
		Specified: true,
		Value:     value,
	}
}

// Float32 represents an optional float32 value.
type Float32 struct {
	Specified bool
	Value     float32
}

// NewString returns a specified String value.
func NewString(value string) String {
	return String{
		Specified: true,
		Value:     value,
	}
}

// String represents an optional string value.
type String struct {
	Specified bool
	Value     string
}

// NewColor returns a specified Color value.
func NewColor(value ui.Color) Color {
	return Color{
		Specified: true,
		Value:     value,
	}
}

// Color represents an optional ui Color value.
type Color struct {
	Specified bool
	Value     ui.Color
}

// NewSpacing returns a specified Spacing value.
func NewSpacing(value ui.Spacing) Spacing {
	return Spacing{
		Specified: true,
		Value:     value,
	}
}

// Spacing represents an optional ui Spacing value.
type Spacing struct {
	Specified bool
	Value     ui.Spacing
}

// NewSize returns a specified Size value.
func NewSize(value ui.Size) Size {
	return Size{
		Specified: true,
		Value:     value,
	}
}

// Size represents an optional ui Size value.
type Size struct {
	Specified bool
	Value     ui.Size
}
