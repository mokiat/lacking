package mat

// Alignment determines the positioning of child elements
// or text within a Layout or Control.
type Alignment int

const (
	AlignmentDefault Alignment = iota
	AlignmentCenter
	AlignmentLeft
	AlignmentRight
	AlignmentTop
	AlignmentBottom
)

// ClickListener can be used to get notifications about
// click events.
type ClickListener func()

type buttonState = int

const (
	buttonStateUp buttonState = 1 + iota
	buttonStateOver
	buttonStateDown
)
