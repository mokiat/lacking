package app

// GamepadState represents a snapshot state of a gamepad controller.
type GamepadState struct {

	// LeftStickX indicates the amount that the left stick has been
	// moved along the X axis.
	LeftStickX float32

	// LeftStickY indicates the amount that the left stick has been
	// moved along the Y axis.
	LeftStickY float32

	// RightStickX indicates the amount that the right stick has been
	// moved along the X axis.
	RightStickX float32

	// RightStickY indicates the amount that the right stick has been
	// moved along the Y axis.
	RightStickY float32

	// LeftTrigger indicates the amount that the left trigger has been
	// pressed.
	LeftTrigger float32

	// RightTrigger indicates the amount that the right trigger has been
	// pressed.
	RightTrigger float32

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
}
