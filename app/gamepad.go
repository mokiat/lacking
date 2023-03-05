package app

import "time"

// NOTE: As much as I would have liked to include joysticks and racing wheels,
// there just isn't an official standard for them and it would not make sense
// for me to come up with something on my own that I can only test with the
// single joystick and wheel that I own.
//
// Shame...

// Gamepad represents a gamepad type joystick. Only input devides that can
// be mapped according to standard layout will work and have any axis
// and button output.
type Gamepad interface {

	// Connected returns whether this Gamepad is still connected. This is
	// useful for when a Gamepad instance is kept by the user code.
	Connected() bool

	// Supported returns whether the connected device can be mapped to the
	// standard layout:
	// https://w3c.github.io/gamepad/#remapping
	//
	// Devices that are not supported will return zero values for buttons
	// and axes.
	Supported() bool

	// StickDeadzone returns the range around 0.0 that will not be considered
	// valid for stick motion.
	StickDeadzone() float64

	// SetStickDeadzone changes the deadzone range for sticks.
	SetStickDeadzone(deadzone float64)

	// TriggerDeadzone returns the range around 0.0 that will not be considered
	// valid for trigger motion.
	TriggerDeadzone() float64

	// SetTriggerDeadzone changes the deadzone range for triggers.
	SetTriggerDeadzone(deadzone float64)

	// LeftStickX returns the horizontal axis of the left stick.
	LeftStickX() float64

	// LeftStickY returns the vertical axis of the left stick.
	LeftStickY() float64

	// LeftStickButton returns the button represented by pressing
	// on the left stick.
	LeftStickButton() bool

	// RightStickX returns the horizontal axis of the right stick.
	RightStickX() float64

	// RightStickY returns the horizontal axis of the right stick.
	RightStickY() float64

	// RightStickButton returns the button represented by pressing
	// on the right stick.
	RightStickButton() bool

	// LeftTrigger returns the left trigger button.
	LeftTrigger() float64

	// RightTrigger returns the right trigger button.
	RightTrigger() float64

	// LeftBumper returns the left bumper button.
	LeftBumper() bool

	// RightBumper returns the right bumper button.
	RightBumper() bool

	// DpadUpButton returns the up button of the left cluster.
	DpadUpButton() bool

	// DpadDownButton returns the down button of the left cluster.
	DpadDownButton() bool

	// DpadLeftButton returns the left button of the left cluster.
	DpadLeftButton() bool

	// DpadRightButton returns the right button of the left cluster.
	DpadRightButton() bool

	// ActionTopButton returns the up button of the right cluster.
	ActionUpButton() bool

	// ActionDownButton returns the down button of the right cluster.
	ActionDownButton() bool

	// ActionLeftButton returns the left button of the right cluster.
	ActionLeftButton() bool

	// ActionRightButton returns the right button of the right cluster.
	ActionRightButton() bool

	// ForwardButton represents the right button of the center cluster.
	ForwardButton() bool

	// BackButton represents the left button of the center cluster.
	BackButton() bool

	// Pulse causes the Gamepad controller to vibrate with the specified
	// intensity (0.0 to 1.0) for the specified duration.
	//
	// If the device does not have haptic feedback then this method does nothing.
	Pulse(intensity float64, duration time.Duration)
}
