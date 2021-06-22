package standard

import (
	"strings"

	"github.com/mokiat/lacking/ui"
)

// Alignment determines the positioning of child elements
// or text within a Layout or Control.
type Alignment int

const (
	AlignmentCenter Alignment = 1 + iota
	AlignmentLeft
	AlignmentRight
	AlignmentTop
	AlignmentBottom
)

// AlignmentAttribute attempts to parse an Alignment from
// the attribute with the specified name.
func AlignmentAttribute(set ui.AttributeSet, name string) (Alignment, bool) {
	if stringValue, ok := set.StringAttribute(name); ok {
		switch strings.ToLower(stringValue) {
		case "left":
			return AlignmentLeft, true
		case "right":
			return AlignmentRight, true
		case "top":
			return AlignmentTop, true
		case "bottom":
			return AlignmentBottom, true
		case "center", "centre", "middle":
			return AlignmentCenter, true
		}
	}
	return 0, false
}

// Relation determines how a position is determined
// (in relation to what).
type Relation int

const (
	RelationLeft Relation = 1 + iota
	RelationRight
	RelationTop
	RelationBottom
	RelationCenter
)

// RelationAttribute attempts to parse a Relation from
// the attribute with the specified name.
func RelationAttribute(set ui.AttributeSet, name string) (Relation, bool) {
	if value, ok := set.StringAttribute(name); ok {
		switch strings.ToLower(value) {
		case "left":
			return RelationLeft, true
		case "right":
			return RelationRight, true
		case "top":
			return RelationTop, true
		case "bottom":
			return RelationBottom, true
		case "center", "centre", "middle":
			return RelationCenter, true
		}
	}
	return 0, false
}
