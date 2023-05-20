package mat

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
