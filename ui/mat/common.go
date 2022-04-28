package mat

const (
	// AlignmentDefault indicates that the Layout should use its preferred
	// alignment model.
	AlignmentDefault Alignment = iota

	// AlignmentCenter indicates that the Layout should try and center
	// the child.
	AlignmentCenter

	// AlignmentLeft indicates that the Layout should try and position the
	// child to the left.
	AlignmentLeft

	// AlignmentRight indicates that the Layout should try and position the
	// child to the right.
	AlignmentRight

	// AlignmentTop indicates that the Layout should try and position the
	// child to the top.
	AlignmentTop

	// AlignmentBottom indicates that the Layout should try and position the
	// child to the bottom.
	AlignmentBottom
)

// Alignment determines the positioning of child Elements
// or text within a Layout.
type Alignment int

const (
	// ButtonStateUp indicates that the button is in its default state.
	ButtonStateUp ButtonState = iota

	// ButtonStateOver indicates that the cursor is over the button.
	ButtonStateOver

	// ButtonStateDown indicates that the cursor is pressing on the button.
	ButtonStateDown
)

// ButtonState indicates the state of a Button control.
type ButtonState int

// ClickListener can be used to get notifications about click events.
type ClickListener func()

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
