package app

// GamepadState represents a snapshot state of a gamepad controller.
type GamepadState struct {

	// LeftStickX indicates the amount that the left stick has been
	// moved along the X axis.
	LeftStickX float64

	// LeftStickY indicates the amount that the left stick has been
	// moved along the Y axis.
	LeftStickY float64

	// RightStickX indicates the amount that the right stick has been
	// moved along the X axis.
	RightStickX float64

	// RightStickY indicates the amount that the right stick has been
	// moved along the Y axis.
	RightStickY float64

	// LeftTrigger indicates the amount that the left trigger has been
	// pressed.
	LeftTrigger float64

	// RightTrigger indicates the amount that the right trigger has been
	// pressed.
	RightTrigger float64

	// LeftBumper indicates whether the left bumper has been pressed.
	LeftBumper bool

	// RightBumper indicates whether the right bumper has been pressed.
	RightBumper bool

	// TriangleButton indicates whether the triangle button has been pressed.
	TriangleButton bool

	// SquareButton indicates whether the square button has been pressed.
	SquareButton bool

	// CrossButton indicates whether the cross button has been pressed.
	CrossButton bool

	// CircleButton indicates whether the circle button has been pressed.
	CircleButton bool

	// DpadUpButton indicates whether the d-pad up button has been pressed.
	DpadUpButton bool

	// DpadDownButton indicates whether the d-pad down button has been pressed.
	DpadDownButton bool

	// DpadLeftButton indicates whether the d-pad left button has been pressed.
	DpadLeftButton bool

	// DpadRightButton indicates whether the d-pad right button has been pressed.
	DpadRightButton bool
}
