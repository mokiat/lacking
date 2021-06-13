package ui

import (
	"strconv"
	"strings"
)

// AttributeSet represents a set of attributes that can
// be applied to Elements and Controls.
type AttributeSet interface {

	// BoolAttribute attempts to read the attribute with the
	// specified name as a boolean. If an attribute with such name
	// does not exist or is not a boolean the boolean flag would
	// indicate that.
	BoolAttribute(name string) (bool, bool)

	// IntAttribute attempts to read the attribute with the
	// specified name as an integer. If an attribute with such name
	// does not exist or is not an integer the boolean flag would
	// indicate that.
	IntAttribute(name string) (int, bool)

	// FloatAttribute attempts to read the attribute with the
	// specified name as a float. If an attribute with such name
	// does not exist or is not a float the boolean flag would
	// indicate that.
	FloatAttribute(name string) (float32, bool)

	// StringAttribute attempts to read the attribute with the
	// specified name. If an attribute with such name does not exist
	// the boolean flag would indicate that.
	StringAttribute(name string) (string, bool)

	// ColorAttribute attempts to read the attribute with the
	// specified name as a Color. If an attribute with such name
	// does not exist or is not a Color the boolean flag would
	// indicate that.
	// Colors are represented as `#RRGGBBAA`, where the `AA` segment
	// is optional, or as the name of one of the default colors in
	// this package (e.g. `red`, `green`, `blue`, etc.)
	ColorAttribute(name string) (Color, bool)
}

// NewMapAttributeSet creates an AttributeSet off of the specified
// map of string keys and values.
func NewMapAttributeSet(entries map[string]string) *MapAttributeSet {
	lowercaseEntries := make(map[string]string)
	for k, v := range entries {
		lowercaseEntries[strings.ToLower(k)] = strings.ToLower(v)
	}
	return &MapAttributeSet{
		entries: lowercaseEntries,
	}
}

var _ AttributeSet = (*MapAttributeSet)(nil)

// MapAttributeSet is an implementation of AttributeSet that holds
// its data in a map of string keys and values.
type MapAttributeSet struct {
	entries map[string]string
}

// BoolAttribute attempts to read the attribute with the
// specified name as a boolean. If an attribute with such name
// does not exist or is not a boolean the boolean flag would
// indicate that.
func (s *MapAttributeSet) BoolAttribute(name string) (bool, bool) {
	value, ok := s.entries[name]
	if !ok {
		return false, false
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return boolValue, true
}

// IntAttribute attempts to read the attribute with the
// specified name as an integer. If an attribute with such name
// does not exist or is not an integer the boolean flag would
// indicate that.
func (s *MapAttributeSet) IntAttribute(name string) (int, bool) {
	value, ok := s.entries[name]
	if !ok {
		return 0, false
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return int(intValue), true
}

// FloatAttribute attempts to read the attribute with the
// specified name as a float. If an attribute with such name
// does not exist or is not a float the boolean flag would
// indicate that.
func (s *MapAttributeSet) FloatAttribute(name string) (float32, bool) {
	value, ok := s.entries[name]
	if !ok {
		return 0, false
	}
	floatValue, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, false
	}
	return float32(floatValue), true
}

// StringAttribute attempts to read the attribute with the
// specified name. If an attribute with such name does not exist
// the boolean flag would indicate that.
func (s *MapAttributeSet) StringAttribute(name string) (string, bool) {
	value, ok := s.entries[name]
	return value, ok
}

// ColorAttribute attempts to read the attribute with the
// specified name as a Color. If an attribute with such name
// does not exist or is not a Color the boolean flag would
// indicate that.
// Colors are represented as `#RRGGBBAA`, where the `AA` segment
// is optional, or as the name of one of the default colors in
// this package (e.g. `red`, `green`, `blue`, etc.)
func (s *MapAttributeSet) ColorAttribute(name string) (Color, bool) {
	value, ok := s.entries[name]
	if !ok {
		return Color{}, false
	}
	switch {
	case KnownColorName(value):
		return NamedColor(value)
	case strings.HasPrefix(value, "#"):
		return ParseColor(value)
	default:
		return Color{}, false
	}
}
