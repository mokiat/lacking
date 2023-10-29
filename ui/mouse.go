package ui

import (
	"fmt"

	"github.com/mokiat/lacking/app"
)

// MouseEvent represents an event related to a mouse action.
type MouseEvent struct {

	// Index indicates which mouse triggered the event. By default
	// the index for a the primary mouse is 0.
	// This is applicable for devices with multiple pointers
	// (mobile) or in case a second mouse is emulated
	// (e.g. with a game controller).
	Index int

	// Action specifies the mouse event type.
	Action MouseAction

	// Button specifies the button for which the event is
	// applicable.
	Button MouseButton

	// X contains the horizontal coordinate of the event.
	X int

	// Y contains the vertical coordinate of the event.
	Y int

	// ScrollX determines the amount of horizontal scroll.
	ScrollX float32

	// ScrollY determines the amount of vertical scroll.
	ScrollY float32

	// Modifiers contains active key modifiers.
	Modifiers KeyModifierSet

	// Payload contains the data that was dropped.
	Payload interface{}
}

// Position is a helper function that returns the position of the event
// based off of the X and Y coordinates.
func (e MouseEvent) Position() Position {
	return NewPosition(e.X, e.Y)
}

// String returns a string representation for this mouse event.
func (e MouseEvent) String() string {
	return fmt.Sprintf("(%d,%s,%s,(%d,%d),(%f,%f),%s)",
		e.Index,
		e.Action,
		e.Button,
		e.X,
		e.Y,
		e.ScrollX,
		e.ScrollY,
		e.Modifiers,
	)
}

const (
	MouseActionDown   MouseAction = MouseAction(app.MouseActionDown)
	MouseActionUp     MouseAction = MouseAction(app.MouseActionUp)
	MouseActionMove   MouseAction = MouseAction(app.MouseActionMove)
	MouseActionEnter  MouseAction = MouseAction(app.MouseActionEnter)
	MouseActionLeave  MouseAction = MouseAction(app.MouseActionLeave)
	MouseActionScroll MouseAction = MouseAction(app.MouseActionScroll)
	MouseActionDrop   MouseAction = MouseAction(app.MouseActionDrop)
)

// MouseAction represents the type of mouse event.
type MouseAction int

// String returns a string representation of this event type.
func (a MouseAction) String() string {
	switch a {
	case MouseActionDown:
		return "DOWN"
	case MouseActionUp:
		return "UP"
	case MouseActionMove:
		return "MOVE"
	case MouseActionEnter:
		return "ENTER"
	case MouseActionLeave:
		return "LEAVE"
	case MouseActionDrop:
		return "DROP"
	case MouseActionScroll:
		return "SCROLL"
	default:
		return "UNKNOWN"
	}
}

const (
	MouseButtonLeft   MouseButton = MouseButton(app.MouseButtonLeft)
	MouseButtonMiddle MouseButton = MouseButton(app.MouseButtonMiddle)
	MouseButtonRight  MouseButton = MouseButton(app.MouseButtonRight)
)

// MouseButton represents the mouse button.
type MouseButton int

// String returns a string representation of this button.
func (b MouseButton) String() string {
	switch b {
	case MouseButtonLeft:
		return "LEFT"
	case MouseButtonMiddle:
		return "MIDDLE"
	case MouseButtonRight:
		return "RIGHT"
	default:
		return "UNKNOWN"
	}
}

// FilepathPayload is a type of Payload that occurs when files
// have been dragged and dropped into the window.
type FilepathPayload = app.FilepathPayload
