package app

import (
	"fmt"
	"time"
)

var (
	// GamepadMinEventInterval is the minimum interval between
	// gamepad event processing. This is to ensure that even if there are no
	// native events, the gamepad will still produce events.
	//
	// This value can be changed from the UI thread.
	GamepadMinEventInterval = 30 * time.Millisecond

	// GamepadRepeatDelay is the initial delay before a held down
	// gamepad button starts repeating.
	//
	// This value can be changed from the UI thread.
	GamepadRepeatDelay = 600 * time.Millisecond

	// GamepadRepeatInterval is the interval between repeated
	// gamepad button events after the initial delay.
	//
	// This value can be changed from the UI thread.
	GamepadRepeatInterval = 100 * time.Millisecond
)

// GamepadEvent represents an event related to a gamepad action.
type GamepadEvent struct {

	// Index indicates which gamepad triggered the event. By default
	// the index for the primary gamepad is 0.
	Index int

	// Gamepad is a reference to the gamepad that triggered the event.
	Gamepad Gamepad

	// Action indicates the action performed with the gamepad.
	Action GamepadAction

	// Button specifies the button for which the event is applicable.
	Button GamepadButton

	// Stick specifies the stick for which the event is applicable.
	Stick GamepadStick

	// X specifies the horizontal position of the stick.
	X float64

	// Y specifies the vertical position of the stick.
	Y float64
}

// String returns a string representation of this event.
func (s GamepadEvent) String() string {
	return fmt.Sprintf("(%d,%s,%s,%s,(%f,%f))",
		s.Index,
		s.Action,
		s.Button,
		s.Stick,
		s.X,
		s.Y,
	)
}

const (
	// GamepadActionNone indicates that no action occurred. This is an unlikely
	// value but allows for more custom event situations.
	GamepadActionNone GamepadAction = iota

	// GamepadActionConnected indicates that the gamepad was connected.
	GamepadActionConnected

	// GamepadActionDisconnected indicates that the gamepad was disconnected.
	GamepadActionDisconnected

	// GamepadActionButtonDown indicates that a gamepad button was pressed down.
	GamepadActionButtonDown

	// GamepadActionButtonUp indicates that a gamepad button was released.
	GamepadActionButtonUp

	// GamepadActionButtonRepeat indicates that a gamepad button is being held
	// down and is repeating.
	GamepadActionButtonRepeat

	// GamepadActionStickMove indicates that a gamepad stick or trigger was moved.
	GamepadActionStickMove
)

// GamepadAction is used to specify the type of gamepad action that occurred.
type GamepadAction int

// String returns a string representation of this event type.
func (s GamepadAction) String() string {
	switch s {
	case GamepadActionNone:
		return "None"
	case GamepadActionConnected:
		return "Connected"
	case GamepadActionDisconnected:
		return "Disconnected"
	case GamepadActionButtonDown:
		return "ButtonDown"
	case GamepadActionButtonUp:
		return "ButtonUp"
	case GamepadActionButtonRepeat:
		return "ButtonRepeat"
	case GamepadActionStickMove:
		return "StickMove"
	default:
		return "Unknown"
	}
}

const (
	// GamepadButtonNone indicates that no button is associated with the event.
	GamepadButtonNone GamepadButton = iota

	// GamepadButtonLeftStick indicates the button represented by pressing
	// on the left stick.
	GamepadButtonLeftStick

	// GamepadButtonRightStick indicates the button represented by pressing
	// on the right stick.
	GamepadButtonRightStick

	// GamepadButtonLeftTrigger indicates the left trigger viewed as a button.
	GamepadButtonLeftTrigger

	// GamepadButtonRightTrigger indicates the right trigger viewed as a button.
	GamepadButtonRightTrigger

	// GamepadButtonLeftBumper indicates the left bumper button.
	GamepadButtonLeftBumper

	// GamepadButtonRightBumper indicates the right bumper button.
	GamepadButtonRightBumper

	// GamepadButtonDpadUp indicates the up button of the left cluster.
	GamepadButtonDpadUp

	// GamepadButtonDpadDown indicates the down button of the left cluster.
	GamepadButtonDpadDown

	// GamepadButtonDpadLeft indicates the left button of the left cluster.
	GamepadButtonDpadLeft

	// GamepadButtonDpadRight indicates the right button of the left cluster.
	GamepadButtonDpadRight

	// GamepadButtonActionUp indicates the up button of the right cluster.
	GamepadButtonActionUp

	// GamepadButtonActionDown indicates the down button of the right cluster.
	GamepadButtonActionDown

	// GamepadButtonActionLeft indicates the left button of the right cluster.
	GamepadButtonActionLeft

	// GamepadButtonActionRight indicates the right button of the right cluster.
	GamepadButtonActionRight

	// GamepadButtonForward indicates the right button of the center cluster.
	GamepadButtonForward

	// GamepadButtonBack indicates the left button of the center cluster.
	GamepadButtonBack

	// GamepadButtonLeftStickUp indicates the up direction on the left stick.
	GamepadButtonLeftStickUp

	// GamepadButtonLeftStickDown indicates the down direction on the left stick.
	GamepadButtonLeftStickDown

	// GamepadButtonLeftStickLeft indicates the left direction on the left stick.
	GamepadButtonLeftStickLeft

	// GamepadButtonLeftStickRight indicates the right direction on the left stick.
	GamepadButtonLeftStickRight

	// GamepadButtonRightStickUp indicates the up direction on the right stick.
	GamepadButtonRightStickUp

	// GamepadButtonRightStickDown indicates the down direction on the right stick.
	GamepadButtonRightStickDown

	// GamepadButtonRightStickLeft indicates the left direction on the right stick.
	GamepadButtonRightStickLeft

	// GamepadButtonRightStickRight indicates the right direction on the right stick.
	GamepadButtonRightStickRight

	// GamepadButtonCount is the total number of gamepad buttons enums.
	GamepadButtonCount
)

// GamepadButton represents the gamepad button.
type GamepadButton int

func (b GamepadButton) String() string {
	switch b {
	case GamepadButtonNone:
		return "None"
	case GamepadButtonLeftStick:
		return "LeftStick"
	case GamepadButtonRightStick:
		return "RightStick"
	case GamepadButtonLeftTrigger:
		return "LeftTrigger"
	case GamepadButtonRightTrigger:
		return "RightTrigger"
	case GamepadButtonLeftBumper:
		return "LeftBumper"
	case GamepadButtonRightBumper:
		return "RightBumper"
	case GamepadButtonDpadUp:
		return "DpadUp"
	case GamepadButtonDpadDown:
		return "DpadDown"
	case GamepadButtonDpadLeft:
		return "DpadLeft"
	case GamepadButtonDpadRight:
		return "DpadRight"
	case GamepadButtonActionUp:
		return "ActionUp"
	case GamepadButtonActionDown:
		return "ActionDown"
	case GamepadButtonActionLeft:
		return "ActionLeft"
	case GamepadButtonActionRight:
		return "ActionRight"
	case GamepadButtonForward:
		return "Forward"
	case GamepadButtonBack:
		return "Back"
	case GamepadButtonLeftStickUp:
		return "LeftStickUp"
	case GamepadButtonLeftStickDown:
		return "LeftStickDown"
	case GamepadButtonLeftStickLeft:
		return "LeftStickLeft"
	case GamepadButtonLeftStickRight:
		return "LeftStickRight"
	case GamepadButtonRightStickUp:
		return "RightStickUp"
	case GamepadButtonRightStickDown:
		return "RightStickDown"
	case GamepadButtonRightStickLeft:
		return "RightStickLeft"
	case GamepadButtonRightStickRight:
		return "RightStickRight"
	default:
		return "Unknown"
	}
}

const (
	// GamepadStickNone indicates that no axis is associated with the event.
	GamepadStickNone GamepadStick = iota

	// GamepadStickLeft indicates the left stick.
	GamepadStickLeft

	// GamepadStickRight indicates the right stick.
	GamepadStickRight

	// GamepadStickLeftTrigger indicates the left trigger. Only the Y value
	// is applicable.
	GamepadStickLeftTrigger

	// GamepadStickRightTrigger indicates the right trigger. Only the Y value
	// is applicable.
	GamepadStickRightTrigger

	// GamepadStickCount is the total number of gamepad stick enums.
	GamepadStickCount
)

// GamepadStick is used to specify a particular stick on the gamepad.
type GamepadStick int

// String returns a string representation of this stick.
func (s GamepadStick) String() string {
	switch s {
	case GamepadStickNone:
		return "None"
	case GamepadStickLeft:
		return "Left"
	case GamepadStickRight:
		return "Right"
	case GamepadStickLeftTrigger:
		return "LeftTrigger"
	case GamepadStickRightTrigger:
		return "RightTrigger"
	default:
		return "Unknown"
	}
}

// Gamepad represents a gamepad-type joystick. Only input devices that can
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

	// RightStickY returns the vertical axis of the right stick.
	RightStickY() float64

	// RightStickButton returns the button represented by pressing
	// on the right stick.
	RightStickButton() bool

	// LeftTrigger returns the analog value of the left trigger.
	LeftTrigger() float64

	// RightTrigger returns the analog value of the right trigger.
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

	// ActionUpButton returns the up button of the right cluster.
	ActionUpButton() bool

	// ActionDownButton returns the down button of the right cluster.
	ActionDownButton() bool

	// ActionLeftButton returns the left button of the right cluster.
	ActionLeftButton() bool

	// ActionRightButton returns the right button of the right cluster.
	ActionRightButton() bool

	// ForwardButton returns the right button of the center cluster.
	ForwardButton() bool

	// BackButton returns the left button of the center cluster.
	BackButton() bool

	// Pulse causes the Gamepad controller to vibrate with the specified
	// intensity (0.0 to 1.0) for the specified duration.
	//
	// If the device does not have haptic feedback or if this API implementation
	// does not support it then this method does nothing.
	Pulse(intensity float64, duration time.Duration)
}
