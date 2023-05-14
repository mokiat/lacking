package layout

// HorizontalAlignment determines the horizontal relative positioning of child
// elements within a Layout or text within a component.
type HorizontalAlignment int8

const (
	// HorizontalAlignmentDefault indicates that the receiver should use its
	// preferred alignment model.
	HorizontalAlignmentDefault HorizontalAlignment = iota

	// HorizontalAlignmentCenter indicates that the receiver should try and center
	// the content horizontally.
	HorizontalAlignmentCenter

	// HorizontalAlignmentLeft indicates that the receiver should try and position
	// the content to the left.
	HorizontalAlignmentLeft

	// HorizontalAlignmentRight indicates that the receiver should try and
	// position the content to the right.
	HorizontalAlignmentRight
)

// VerticalAlignment determines the vertical relative positioning of child
// elements within a Layout or text within a component.
type VerticalAlignment int8

const (
	// VerticalAlignmentDefault indicates that the receiver should use its
	// preferred alignment model.
	VerticalAlignmentDefault VerticalAlignment = iota

	// VerticalAlignmentCenter indicates that the receiver should try and center
	// the content vertically.
	VerticalAlignmentCenter

	// VerticalAlignmentTop indicates that the receiver should try and position
	// the content to the top.
	VerticalAlignmentTop

	// VerticalAlignmentBottom indicates that the receiver should try and position
	// the content to the bottom.
	VerticalAlignmentBottom
)
