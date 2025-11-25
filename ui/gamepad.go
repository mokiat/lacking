package ui

import (
	"fmt"

	"github.com/mokiat/lacking/app"
)

// GamepadEvent is used to propagate events related to gamepad
// actions.
type GamepadEvent struct {

	// Index indicates which gamepad triggered the event. By default
	// the index for a the primary gamepad is 0.
	Index int

	// Gamepad is a reference to the gamepad that triggered the event.
	Gamepad app.Gamepad

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
	GamepadActionNone         GamepadAction = GamepadAction(app.GamepadActionNone)
	GamepadActionConnected    GamepadAction = GamepadAction(app.GamepadActionConnected)
	GamepadActionDisconnected GamepadAction = GamepadAction(app.GamepadActionDisconnected)
	GamepadActionButtonDown   GamepadAction = GamepadAction(app.GamepadActionButtonDown)
	GamepadActionButtonUp     GamepadAction = GamepadAction(app.GamepadActionButtonUp)
	GamepadActionButtonRepeat GamepadAction = GamepadAction(app.GamepadActionButtonRepeat)
	GamepadActionStickMove    GamepadAction = GamepadAction(app.GamepadActionStickMove)
)

// GamepadAction is used to specify the type of gamepad
// action that occurred.
type GamepadAction int

// String returns a string representation of this event type,
func (a GamepadAction) String() string {
	return app.GamepadAction(a).String()
}

const (
	GamepadButtonNone            GamepadButton = GamepadButton(app.GamepadButtonNone)
	GamepadButtonLeftStick       GamepadButton = GamepadButton(app.GamepadButtonLeftStick)
	GamepadButtonRightStick      GamepadButton = GamepadButton(app.GamepadButtonRightStick)
	GamepadButtonLeftTrigger     GamepadButton = GamepadButton(app.GamepadButtonLeftTrigger)
	GamepadButtonRightTrigger    GamepadButton = GamepadButton(app.GamepadButtonRightTrigger)
	GamepadButtonLeftBumper      GamepadButton = GamepadButton(app.GamepadButtonLeftBumper)
	GamepadButtonRightBumper     GamepadButton = GamepadButton(app.GamepadButtonRightBumper)
	GamepadButtonDpadUp          GamepadButton = GamepadButton(app.GamepadButtonDpadUp)
	GamepadButtonDpadDown        GamepadButton = GamepadButton(app.GamepadButtonDpadDown)
	GamepadButtonDpadLeft        GamepadButton = GamepadButton(app.GamepadButtonDpadLeft)
	GamepadButtonDpadRight       GamepadButton = GamepadButton(app.GamepadButtonDpadRight)
	GamepadButtonActionUp        GamepadButton = GamepadButton(app.GamepadButtonActionUp)
	GamepadButtonActionDown      GamepadButton = GamepadButton(app.GamepadButtonActionDown)
	GamepadButtonActionLeft      GamepadButton = GamepadButton(app.GamepadButtonActionLeft)
	GamepadButtonActionRight     GamepadButton = GamepadButton(app.GamepadButtonActionRight)
	GamepadButtonForward         GamepadButton = GamepadButton(app.GamepadButtonForward)
	GamepadButtonBack            GamepadButton = GamepadButton(app.GamepadButtonBack)
	GamepadButtonLeftStickUp     GamepadButton = GamepadButton(app.GamepadButtonLeftStickUp)
	GamepadButtonLeftStickDown   GamepadButton = GamepadButton(app.GamepadButtonLeftStickDown)
	GamepadButtonLeftStickLeft   GamepadButton = GamepadButton(app.GamepadButtonLeftStickLeft)
	GamepadButtonLeftStickRight  GamepadButton = GamepadButton(app.GamepadButtonLeftStickRight)
	GamepadButtonRightStickUp    GamepadButton = GamepadButton(app.GamepadButtonRightStickUp)
	GamepadButtonRightStickDown  GamepadButton = GamepadButton(app.GamepadButtonRightStickDown)
	GamepadButtonRightStickLeft  GamepadButton = GamepadButton(app.GamepadButtonRightStickLeft)
	GamepadButtonRightStickRight GamepadButton = GamepadButton(app.GamepadButtonRightStickRight)
)

// GamepadButton represents the gamepad button.
type GamepadButton int

// String returns a string representation of this button.
func (b GamepadButton) String() string {
	return app.GamepadButton(b).String()
}

const (
	GamepadStickNone         GamepadStick = GamepadStick(app.GamepadStickNone)
	GamepadStickLeft         GamepadStick = GamepadStick(app.GamepadStickLeft)
	GamepadStickRight        GamepadStick = GamepadStick(app.GamepadStickRight)
	GamepadStickLeftTrigger  GamepadStick = GamepadStick(app.GamepadStickLeftTrigger)
	GamepadStickRightTrigger GamepadStick = GamepadStick(app.GamepadStickRightTrigger)
)

// GamepadStick is used to specify a particular stick on the gamepad.
type GamepadStick int

// String returns a string representation of this stick.
func (s GamepadStick) String() string {
	return app.GamepadStick(s).String()
}
